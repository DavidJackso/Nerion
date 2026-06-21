package thttp

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"nerion/internal/entity"
	authmw "nerion/internal/middleware"
)

func (s *Server) apiKeyRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(authmw.Auth(s.jwtManager))
	r.Use(authmw.LoadSpace(s.spaceRepo))
	r.Use(authmw.RequireSpaceRole(entity.SpaceMemberRoleAdmin, s.memberRepo))

	r.Get("/", s.listAPIKeys)
	r.Post("/", s.createAPIKey)
	r.Delete("/{id}", s.revokeAPIKey)

	return r
}

type apiKeyResponse struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	Scope      string     `json:"scope"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

type apiKeyCreateResponse struct {
	apiKeyResponse
	Key string `json:"key"`
}

func toAPIKeyResponse(k *entity.APIKey) apiKeyResponse {
	return apiKeyResponse{
		ID:         k.ID,
		Name:       k.Name,
		Prefix:     k.KeyPrefix,
		Scope:      k.Scope,
		Status:     k.Status(),
		CreatedAt:  k.CreatedAt,
		LastUsedAt: k.LastUsedAt,
	}
}

func (s *Server) listAPIKeys(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())
	keys, err := s.apiKeyService.List(r.Context(), space.Slug, claims.UserID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	resp := make([]apiKeyResponse, 0, len(keys))
	for _, k := range keys {
		resp = append(resp, toAPIKeyResponse(k))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) createAPIKey(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())

	key, fullKey, err := s.apiKeyService.Create(r.Context(), space.Slug, req.Name, req.Scope, claims.UserID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	s.auditForSpace(space.ID, claims.UserID, "api_key.create", "api_key", i64str(key.ID), map[string]any{"name": key.Name, "scope": key.Scope})
	writeJSON(w, http.StatusCreated, apiKeyCreateResponse{
		apiKeyResponse: toAPIKeyResponse(key),
		Key:            fullKey,
	})
}

func (s *Server) revokeAPIKey(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректный ID"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())

	if err := s.apiKeyService.Revoke(r.Context(), space.Slug, id, claims.UserID); err != nil {
		s.writeError(w, r, err)
		return
	}
	s.auditForSpace(space.ID, claims.UserID, "api_key.revoke", "api_key", i64str(id), nil)
	w.WriteHeader(http.StatusNoContent)
}
