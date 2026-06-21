package thttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	authmw "go-template/internal/middleware"
)

func (s *Server) userRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(authmw.Auth(s.jwtManager))

	r.Get("/{id}", s.getUser)

	r.Group(func(r chi.Router) {
		r.Use(authmw.RequireRole("admin"))
		r.Get("/", s.listUsers)
		r.Post("/", s.createUser)
	})

	return r
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	limit, offset := 50, 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	users, err := s.userService.ListUsers(r.Context(), limit, offset)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	resp := make([]userResponse, len(users))
	for i, u := range users {
		resp[i] = toUserResponse(u)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, err := s.userService.CreateUser(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		s.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toUserResponse(user))
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	claims, ok := authmw.ClaimsFrom(r.Context())
	if !ok || (claims.Role != "admin" && claims.UserID != id) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	user, err := s.userService.GetUser(r.Context(), id)
	if err != nil {
		s.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}
