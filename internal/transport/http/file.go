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

	"github.com/go-chi/chi/v5"

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
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize+1<<20)
	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Файл слишком большой или некорректный запрос"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Файл не найден в запросе"))
		return
	}
	defer file.Close()

	if header.Size > maxFileSize {
		writeJSON(w, http.StatusRequestEntityTooLarge, errBody("file_too_large", "Файл превышает 50 МБ"))
		return
	}

	// Detect content type from actual bytes — never trust client-supplied header.
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])

	if !allowedContentTypes[contentType] {
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

	if err := s.storage.Upload(r.Context(), key, data, header.Size, contentType); err != nil {
		s.writeError(w, r, err)
		return
	}

	ttl := s.presignTTL
	url, err := s.storage.PresignedURL(r.Context(), key, ttl)
	if err != nil {
		s.writeError(w, r, err)
		return
	}

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
		} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
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
