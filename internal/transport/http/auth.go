package thttp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func (s *Server) authRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/login", s.login)
	return r
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, err := s.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		s.writeError(w, r, err)
		return
	}

	token, err := s.jwtManager.Generate(user.ID, string(user.Role))
	if err != nil {
		s.logger.Error("generate token failed",
			"request_id", chimw.GetReqID(r.Context()),
			"err", err,
		)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token})
}
