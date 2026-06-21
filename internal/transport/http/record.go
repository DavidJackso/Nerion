package thttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	authmw "nerion/internal/middleware"
	"nerion/internal/entity"
)

func (s *Server) recordRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(authmw.Auth(s.jwtManager))

	r.Get("/spaces/{slug}/tables/{table}/records", s.listRecords)
	r.Post("/spaces/{slug}/tables/{table}/records", s.createRecord)
	r.Get("/spaces/{slug}/tables/{table}/records/{id}", s.getRecord)
	r.Put("/spaces/{slug}/tables/{table}/records/{id}", s.updateRecord)
	r.Delete("/spaces/{slug}/tables/{table}/records/{id}", s.deleteRecord)

	return r
}

func (s *Server) listRecords(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	params := entity.ListParams{
		Limit:   limit,
		Offset:  offset,
		SortBy:  q.Get("sort_by"),
		SortDir: q.Get("sort_dir"),
		Search:  q.Get("search"),
	}

	records, total, err := s.recordService.List(r.Context(),
		chi.URLParam(r, "slug"), chi.URLParam(r, "table"),
		claims.UserID, params)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if records == nil {
		records = []map[string]any{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": records,
		"meta": map[string]any{"total": total, "limit": params.Limit, "offset": params.Offset},
	})
}

func (s *Server) createRecord(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<20)
	var data map[string]any
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	rec, err := s.recordService.Create(r.Context(),
		chi.URLParam(r, "slug"), chi.URLParam(r, "table"),
		claims.UserID, data)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	space, _ := authmw.SpaceFrom(r.Context())
	s.auditForSpace(space.ID, claims.UserID, "record.create", chi.URLParam(r, "table"), anyToStr(rec["id"]), nil)
	writeJSON(w, http.StatusCreated, rec)
}

func (s *Server) getRecord(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректный ID"))
		return
	}
	rec, err := s.recordService.GetByID(r.Context(),
		chi.URLParam(r, "slug"), chi.URLParam(r, "table"),
		claims.UserID, id)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (s *Server) updateRecord(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<20)
	var data map[string]any
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректный ID"))
		return
	}
	rec, err := s.recordService.Update(r.Context(),
		chi.URLParam(r, "slug"), chi.URLParam(r, "table"),
		claims.UserID, id, data)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (s *Server) deleteRecord(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректный ID"))
		return
	}
	if err := s.recordService.Delete(r.Context(),
		chi.URLParam(r, "slug"), chi.URLParam(r, "table"),
		claims.UserID, id); err != nil {
		s.writeError(w, r, err)
		return
	}
	space, _ := authmw.SpaceFrom(r.Context())
	s.auditForSpace(space.ID, claims.UserID, "record.delete", chi.URLParam(r, "table"), i64str(id), nil)
	w.WriteHeader(http.StatusNoContent)
}
