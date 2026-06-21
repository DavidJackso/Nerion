package thttp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"nerion/internal/entity"
	authmw "nerion/internal/middleware"
)

// audit fires an audit log entry asynchronously so it never blocks a request.
func (s *Server) audit(spaceID *int64, userID *int64, action, entityType, entityID string, meta map[string]any) {
	e := &entity.AuditEntry{
		SpaceID:    spaceID,
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Meta:       meta,
	}
	go func() {
		_ = s.auditRepo.Log(context.Background(), e)
	}()
}

// auditForSpace is a convenience wrapper when space and user are known.
func (s *Server) auditForSpace(spaceID, userID int64, action, entityType, entityID string, meta map[string]any) {
	s.audit(&spaceID, &userID, action, entityType, entityID, meta)
}

func (s *Server) listAuditLog(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())

	// Admin only
	role, err := s.memberRepo.GetRole(r.Context(), space.ID, claims.UserID)
	if err != nil || role != entity.SpaceMemberRoleAdmin {
		writeJSON(w, http.StatusForbidden, errBody("forbidden", "Требуется роль admin"))
		return
	}

	limit, offset := 50, 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	entries, err := s.auditRepo.List(r.Context(), space.ID, limit, offset)
	if err != nil {
		s.writeError(w, r, err)
		return
	}

	resp := make([]auditEntryResponse, len(entries))
	for i, e := range entries {
		resp[i] = toAuditResponse(e)
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": resp, "meta": map[string]any{"limit": limit, "offset": offset}})
}

func (s *Server) auditMetrics(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())

	role, err := s.memberRepo.GetRole(r.Context(), space.ID, claims.UserID)
	if err != nil || role != entity.SpaceMemberRoleAdmin {
		writeJSON(w, http.StatusForbidden, errBody("forbidden", "Требуется роль admin"))
		return
	}

	apiKeyCounts, err := s.auditRepo.CountByAction(r.Context(), space.ID, "api_key.request")
	if err != nil {
		s.writeError(w, r, err)
		return
	}

	// PDF metrics from audit_log
	pdfCounts, err := s.auditRepo.CountByAction(r.Context(), space.ID, "pdf.generate")
	if err != nil {
		s.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"api_key_requests": apiKeyCounts,
		"pdf_jobs":         pdfCounts,
	})
}

func (s *Server) auditRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(authmw.Auth(s.jwtManager))
	r.Use(authmw.LoadSpace(s.spaceRepo))
	r.Get("/", s.listAuditLog)
	r.Get("/metrics", s.auditMetrics)
	return r
}

type auditEntryResponse struct {
	ID         int64          `json:"id"`
	SpaceID    *int64         `json:"space_id,omitempty"`
	UserID     *int64         `json:"user_id,omitempty"`
	Action     string         `json:"action"`
	EntityType string         `json:"entity_type,omitempty"`
	EntityID   string         `json:"entity_id,omitempty"`
	Meta       map[string]any `json:"meta,omitempty"`
	CreatedAt  string         `json:"created_at"`
}

func toAuditResponse(e *entity.AuditEntry) auditEntryResponse {
	return auditEntryResponse{
		ID:         e.ID,
		SpaceID:    e.SpaceID,
		UserID:     e.UserID,
		Action:     e.Action,
		EntityType: e.EntityType,
		EntityID:   e.EntityID,
		Meta:       e.Meta,
		CreatedAt:  e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func i64str(n int64) string  { return fmt.Sprintf("%d", n) }
func anyToStr(v any) string  { return fmt.Sprintf("%v", v) }
