package thttp

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"nerion/internal/entity"
	authmw "nerion/internal/middleware"
)

func (s *Server) pdfRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(authmw.Auth(s.jwtManager))
	r.Use(authmw.LoadSpace(s.spaceRepo))

	r.Get("/templates", s.listPDFTemplates)
	r.Post("/templates", s.uploadPDFTemplate)
	r.Put("/templates/{id}/mapping", s.savePDFMapping)
	r.Post("/templates/{id}/preview", s.previewPDF)
	r.Post("/generate", s.generatePDF)
	r.Get("/jobs/{jobID}", s.getPDFJob)
	r.Get("/archive", s.listPDFArchive)

	return r
}

func (s *Server) listPDFTemplates(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())

	templates, err := s.pdfService.ListTemplates(r.Context(), space.Slug, claims.UserID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if templates == nil {
		templates = []*entity.PDFTemplate{}
	}
	writeJSON(w, http.StatusOK, templates)
}

func (s *Server) uploadPDFTemplate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 20<<20) // 20 MB
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректный multipart запрос"))
		return
	}

	name := r.FormValue("name")
	if name == "" {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Поле name обязательно"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Файл не найден в запросе"))
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		s.writeError(w, r, err)
		return
	}

	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())

	t, err := s.pdfService.UploadTemplate(r.Context(), space.Slug, name, claims.UserID, data, header.Filename)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (s *Server) savePDFMapping(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректный ID"))
		return
	}

	var req saveMappingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}

	mappings := make([]*entity.PDFMapping, len(req.Mappings))
	for i, m := range req.Mappings {
		mappings[i] = &entity.PDFMapping{
			Placeholder:   m.Placeholder,
			SourceFieldID: m.SourceFieldID,
			Expression:    m.Expression,
		}
	}

	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())

	if err := s.pdfService.SaveMapping(r.Context(), space.Slug, id, claims.UserID, mappings); err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Маппинг сохранён"})
}

func (s *Server) previewPDF(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректный ID"))
		return
	}

	var req pdfPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}

	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())

	result, err := s.pdfService.Preview(r.Context(), space.Slug, id, req.RecordID, req.TableSlug, claims.UserID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) generatePDF(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req pdfGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	if req.TemplateID == 0 {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "template_id обязателен"))
		return
	}

	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())

	job, err := s.pdfService.Generate(r.Context(), space.Slug, req.TemplateID, claims.UserID, req.TableSlug, req.RecordIDs)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	s.auditForSpace(space.ID, claims.UserID, "pdf.generate", "pdf_job", job.ID, map[string]any{"template_id": req.TemplateID, "record_count": len(req.RecordIDs)})
	writeJSON(w, http.StatusCreated, job)
}

func (s *Server) getPDFJob(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())
	jobID := chi.URLParam(r, "jobID")

	job, err := s.pdfService.GetJob(r.Context(), space.Slug, jobID, claims.UserID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) listPDFArchive(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())

	jobs, err := s.pdfService.ListArchive(r.Context(), space.Slug, claims.UserID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if jobs == nil {
		jobs = []*entity.PDFJob{}
	}
	writeJSON(w, http.StatusOK, jobs)
}
