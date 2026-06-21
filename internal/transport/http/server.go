package thttp

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"go-template/internal/domain"
	"go-template/internal/jwtauth"
	authmw "go-template/internal/middleware"
	"go-template/pkg/apierrors"
)

type Server struct {
	userService domain.UserService
	jwtManager  jwtauth.Tokenizer
	logger      *slog.Logger
	router      chi.Router
}

func NewServer(userService domain.UserService, jm jwtauth.Tokenizer, logger *slog.Logger) *Server {
	s := &Server{
		userService: userService,
		jwtManager:  jm,
		logger:      logger,
		router:      chi.NewRouter(),
	}
	s.router.Use(
		chimw.RequestID,
		chimw.RealIP,
		authmw.RequestLogger(logger),
		chimw.Recoverer,
		chimw.CleanPath,
		chimw.Timeout(8*time.Second),
	)
	s.registerRoutes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	s.router.Mount("/auth", s.authRoutes())
	s.router.Mount("/users", s.userRoutes())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) writeError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *apierrors.APIError
	if errors.As(err, &apiErr) {
		writeJSON(w, apiErr.Code, map[string]string{"error": apiErr.Message})
		return
	}
	s.logger.Error("unexpected error",
		"request_id", chimw.GetReqID(r.Context()),
		"err", err,
	)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}
