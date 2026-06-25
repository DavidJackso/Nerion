package thttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	authmw "nerion/internal/middleware"
)

func (s *Server) inviteRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{token}", s.getInviteInfo)
	r.Group(func(r chi.Router) {
		r.Use(authmw.Auth(s.jwtManager))
		r.Post("/{token}/accept", s.acceptInvite)
	})
	return r
}

func (s *Server) getInviteInfo(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	inv, err := s.memberService.GetInviteInfo(r.Context(), token)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, inviteInfoResponse{
		SpaceID:   inv.SpaceID,
		SpaceName: inv.SpaceName,
		Email:     inv.Email,
	})
}

func (s *Server) acceptInvite(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	claims, _ := authmw.ClaimsFrom(r.Context())
	if err := s.memberService.AcceptInvite(r.Context(), token, claims.UserID); err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Вы успешно добавлены в пространство"})
}
