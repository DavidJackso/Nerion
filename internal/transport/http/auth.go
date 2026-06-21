package thttp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	authmw "nerion/internal/middleware"
)

func (s *Server) authRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", s.register)
	r.Post("/login", s.login)
	r.Post("/refresh", s.refresh)
	r.Post("/logout", s.logout)
	r.Post("/password/reset-request", s.requestPasswordReset)
	r.Post("/password/reset", s.resetPassword)
	r.Get("/verify", s.verifyEmail)
	return r
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	_, err := s.authService.Register(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "Проверьте почту для подтверждения"})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	accessToken, refreshToken, err := s.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if claims, err2 := s.jwtManager.Parse(accessToken); err2 == nil {
		uid := claims.UserID
		s.audit(nil, &uid, "auth.login", "user", i64str(uid), map[string]any{"email": req.Email})
	}
	writeJSON(w, http.StatusOK, loginResponse{AccessToken: accessToken, RefreshToken: refreshToken})
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	accessToken, newRefresh, err := s.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, refreshResponse{AccessToken: accessToken, RefreshToken: newRefresh})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	if err := s.authService.Logout(r.Context(), req.RefreshToken); err != nil {
		s.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req resetRequestReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	_ = s.authService.RequestPasswordReset(r.Context(), req.Email)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Если email зарегистрирован, ссылка отправлена"})
}

func (s *Server) resetPassword(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req resetPasswordReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	if err := s.authService.ResetPassword(r.Context(), req.Token, req.Password); err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Пароль успешно изменён"})
}

func (s *Server) verifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Отсутствует token"))
		return
	}
	if err := s.authService.VerifyEmail(r.Context(), token); err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Email подтверждён"})
}

func (s *Server) getMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmw.ClaimsFrom(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized", "Необходима авторизация"))
		return
	}
	user, err := s.authService.GetMe(r.Context(), claims.UserID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func errBody(code, message string) errorResponse {
	return errorResponse{Error: errorBody{Code: code, Message: message}}
}
