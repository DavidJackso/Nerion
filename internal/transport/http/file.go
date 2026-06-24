package thttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	authmw "nerion/internal/middleware"
)

const maxFileSize = 50 << 20 // 50 MB

// SVG excluded: can contain JavaScript → XSS risk when served from same origin.
var allowedContentTypes = map[string]bool{
	"image/jpeg":       true,
	"image/png":        true,
	"image/gif":        true,
	"image/webp":       true,
	"application/pdf":  true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"text/csv":   true,
	"text/plain": true,
}

func (s *Server) fileRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(authmw.Auth(s.jwtManager))
	r.Use(authmw.LoadSpace(s.spaceRepo))

	r.Post("/upload", s.uploadFile)
	r.Get("/presign", s.presignFile)

	return r
}

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	reqID := chimw.GetReqID(r.Context())

	s.logger.Info("uploadFile: request received",
		"request_id", reqID,
		"remote_addr", r.RemoteAddr,
		"content_length", r.ContentLength,
	)

	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize+1<<20)
	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		s.logger.Warn("uploadFile: failed to parse multipart form",
			"request_id", reqID,
			"err", err,
		)
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Файл слишком большой или некорректный запрос"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.logger.Warn("uploadFile: file field missing in form",
			"request_id", reqID,
			"err", err,
		)
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Файл не найден в запросе"))
		return
	}
	defer file.Close()

	s.logger.Debug("uploadFile: file received",
		"request_id", reqID,
		"filename", header.Filename,
		"declared_size", header.Size,
	)

	if header.Size > maxFileSize {
		s.logger.Warn("uploadFile: file exceeds size limit",
			"request_id", reqID,
			"filename", header.Filename,
			"size", header.Size,
			"max_size", maxFileSize,
		)
		writeJSON(w, http.StatusRequestEntityTooLarge, errBody("file_too_large", "Файл превышает 50 МБ"))
		return
	}

	// Detect content type from actual bytes — never trust client-supplied header.
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])

	s.logger.Debug("uploadFile: content type detected",
		"request_id", reqID,
		"filename", header.Filename,
		"content_type", contentType,
	)

	if !allowedContentTypes[contentType] {
		s.logger.Warn("uploadFile: rejected — content type not allowed",
			"request_id", reqID,
			"filename", header.Filename,
			"content_type", contentType,
		)
		writeJSON(w, http.StatusBadRequest, errBody("invalid_file_type",
			fmt.Sprintf("Тип файла не разрешён: %s", contentType)))
		return
	}

	// Seek back to start + rebuffer
	data := io.MultiReader(bytes.NewReader(buf[:n]), file)

	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())
	_ = claims

	ext := filepath.Ext(header.Filename)
	key := fmt.Sprintf("%s/%d_%s%s",
		space.Slug,
		time.Now().UnixMilli(),
		sanitizeFilename(strings.TrimSuffix(header.Filename, ext)),
		ext,
	)

	s.logger.Info("uploadFile: uploading to storage",
		"request_id", reqID,
		"space", space.Slug,
		"key", key,
		"filename", header.Filename,
		"size", header.Size,
		"content_type", contentType,
	)

	if err := s.storage.Upload(r.Context(), key, data, header.Size, contentType); err != nil {
		s.logger.Error("uploadFile: storage upload failed",
			"request_id", reqID,
			"space", space.Slug,
			"key", key,
			"size", header.Size,
			"err", err,
		)
		s.writeError(w, r, err)
		return
	}

	ttl := s.presignTTL

	s.logger.Debug("uploadFile: generating presigned URL",
		"request_id", reqID,
		"key", key,
		"ttl", ttl,
	)

	url, err := s.storage.PresignedURL(r.Context(), key, ttl)
	if err != nil {
		s.logger.Error("uploadFile: presign failed",
			"request_id", reqID,
			"key", key,
			"err", err,
		)
		s.writeError(w, r, err)
		return
	}

	s.logger.Info("uploadFile: success",
		"request_id", reqID,
		"space", space.Slug,
		"key", key,
		"size", header.Size,
		"content_type", contentType,
	)

	writeJSON(w, http.StatusCreated, map[string]any{
		"key":          key,
		"url":          url,
		"content_type": contentType,
		"size":         header.Size,
		"expires_at":   time.Now().Add(ttl),
	})
}

func (s *Server) presignFile(w http.ResponseWriter, r *http.Request) {
	_, _ = authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())

	key := r.URL.Query().Get("key")
	if key == "" {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Параметр key обязателен"))
		return
	}

	// Security: normalize to block path traversal, then verify space ownership.
	clean := path.Clean(key)
	if clean != key || !strings.HasPrefix(clean, space.Slug+"/") {
		writeJSON(w, http.StatusForbidden, errBody("forbidden", "Ключ не принадлежит данному пространству"))
		return
	}

	url, err := s.storage.PresignedURL(r.Context(), key, s.presignTTL)
	if err != nil {
		s.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"key":        key,
		"url":        url,
		"expires_at": time.Now().Add(s.presignTTL),
	})
}

func sanitizeFilename(name string) string {
	var sb strings.Builder
	for _, r := range name {
		if r == ' ' {
			sb.WriteRune('_')
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			sb.WriteRune(r)
		}
	}
	s := sb.String()
	if len(s) > 64 {
		s = s[:64]
	}
	if s == "" {
		s = "file"
	}
	return s
}
