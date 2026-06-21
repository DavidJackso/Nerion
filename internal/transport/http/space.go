package thttp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	authmw "nerion/internal/middleware"
)

func (s *Server) spaceRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(authmw.Auth(s.jwtManager))

	r.Get("/", s.listSpaces)
	r.Post("/", s.createSpace)

	r.Route("/{slug}", func(r chi.Router) {
		r.Use(authmw.LoadSpace(s.spaceRepo))
		r.Get("/", s.getSpace)
		r.Put("/", s.renameSpace)
		r.Put("/settings", s.updateSpaceSettings)
		r.Delete("/", s.deleteSpace)
	})

	return r
}

func (s *Server) listSpaces(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	spaces, err := s.spaceService.List(r.Context(), claims.UserID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	resp := make([]spaceResponse, 0, len(spaces))
	for _, sp := range spaces {
		tc, _ := s.spaceRepo.TableCount(r.Context(), sp.ID)
		resp = append(resp, toSpaceResponse(sp, tc))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) createSpace(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createSpaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	sp, err := s.spaceService.Create(r.Context(), claims.UserID, req.Name, req.Slug)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toSpaceResponse(sp, 0))
}

func (s *Server) getSpace(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	slug := chi.URLParam(r, "slug")
	sp, err := s.spaceService.Get(r.Context(), claims.UserID, slug)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	tc, _ := s.spaceRepo.TableCount(r.Context(), sp.ID)
	writeJSON(w, http.StatusOK, toSpaceResponse(sp, tc))
}

func (s *Server) renameSpace(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req renameSpaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	slug := chi.URLParam(r, "slug")
	if err := s.spaceService.Rename(r.Context(), claims.UserID, slug, req.Name); err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Переименовано"})
}

// updateSpaceSettings is the canonical settings endpoint (Admin only).
// Currently supports renaming; extensible for future settings (description, icon, etc.).
func (s *Server) updateSpaceSettings(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	slug := chi.URLParam(r, "slug")
	if err := s.spaceService.Rename(r.Context(), claims.UserID, slug, req.Name); err != nil {
		s.writeError(w, r, err)
		return
	}
	space, err := s.spaceService.Get(r.Context(), claims.UserID, slug)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	tc, _ := s.spaceRepo.TableCount(r.Context(), space.ID)
	writeJSON(w, http.StatusOK, toSpaceResponse(space, tc))
}

func (s *Server) deleteSpace(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req deleteSpaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	slug := chi.URLParam(r, "slug")
	if err := s.spaceService.Delete(r.Context(), claims.UserID, slug, req.ConfirmName); err != nil {
		s.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
