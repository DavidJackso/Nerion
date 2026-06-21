package thttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"nerion/internal/entity"
	authmw "nerion/internal/middleware"
)

func (s *Server) memberRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(authmw.Auth(s.jwtManager))
	r.Use(authmw.LoadSpace(s.spaceRepo))

	r.Get("/", s.listMembers)
	r.Post("/invite", s.inviteMember)

	r.Group(func(r chi.Router) {
		r.Use(authmw.RequireSpaceRole(entity.SpaceMemberRoleAdmin, s.memberRepo))
		r.Put("/{userID}/role", s.changeMemberRole)
		r.Delete("/{userID}", s.removeMember)
	})

	return r
}

func (s *Server) listMembers(w http.ResponseWriter, r *http.Request) {
	space, _ := authmw.SpaceFrom(r.Context())
	members, err := s.memberService.ListMembers(r.Context(), space.ID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	resp := make([]memberResponse, 0, len(members))
	for _, m := range members {
		resp = append(resp, memberResponse{
			UserID:    m.UserID,
			UserName:  m.UserName,
			UserEmail: m.UserEmail,
			Role:      string(m.Role),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) inviteMember(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req inviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())
	if err := s.memberService.Invite(r.Context(), space.ID, claims.UserID, req.Email); err != nil {
		s.writeError(w, r, err)
		return
	}
	s.auditForSpace(space.ID, claims.UserID, "member.invite", "user", "", map[string]any{"email": req.Email})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Приглашение отправлено"})
}

func (s *Server) changeMemberRole(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req changeMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	targetID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректный userID"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())
	role := entity.SpaceMemberRole(req.Role)
	if role != entity.SpaceMemberRoleAdmin && role != entity.SpaceMemberRoleMember {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Роль должна быть admin или member"))
		return
	}
	if err := s.memberService.ChangeRole(r.Context(), space.ID, claims.UserID, targetID, role); err != nil {
		s.writeError(w, r, err)
		return
	}
	s.auditForSpace(space.ID, claims.UserID, "member.role_change", "user", i64str(targetID), map[string]any{"role": req.Role})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Роль изменена"})
}

func (s *Server) removeMember(w http.ResponseWriter, r *http.Request) {
	targetID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректный userID"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())
	if err := s.memberService.RemoveMember(r.Context(), space.ID, claims.UserID, targetID); err != nil {
		s.writeError(w, r, err)
		return
	}
	s.auditForSpace(space.ID, claims.UserID, "member.remove", "user", i64str(targetID), nil)
	w.WriteHeader(http.StatusNoContent)
}
