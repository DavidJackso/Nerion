package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"nerion/internal/domain"
	"nerion/internal/entity"
)

type spaceContextKey string

const (
	spaceKey      spaceContextKey = "space"
	spaceMemberKey spaceContextKey = "space_role"
)

// LoadSpace resolves the space from the {slug} URL param and stores it in context.
func LoadSpace(spaceRepo domain.SpaceRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slug := chi.URLParam(r, "slug")
			space, err := spaceRepo.GetBySlug(r.Context(), slug)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]string{"code": "not_found", "message": "Пространство не найдено"},
				})
				return
			}
			ctx := context.WithValue(r.Context(), spaceKey, space)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireSpaceRole checks that the authenticated user has at least the given role in the loaded space.
// Must be used after Auth and LoadSpace middlewares.
func RequireSpaceRole(minRole entity.SpaceMemberRole, memberRepo domain.SpaceMemberRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			space, ok := SpaceFrom(r.Context())
			if !ok {
				writeSpaceErr(w, http.StatusInternalServerError, "internal_error", "Пространство не загружено")
				return
			}
			claims, ok := ClaimsFrom(r.Context())
			if !ok {
				writeSpaceErr(w, http.StatusUnauthorized, "unauthorized", "Необходима авторизация")
				return
			}
			role, err := memberRepo.GetRole(r.Context(), space.ID, claims.UserID)
			if err != nil {
				writeSpaceErr(w, http.StatusForbidden, "forbidden", "Доступ запрещён")
				return
			}
			if minRole == entity.SpaceMemberRoleAdmin && role != entity.SpaceMemberRoleAdmin {
				writeSpaceErr(w, http.StatusForbidden, "forbidden", "Требуются права администратора")
				return
			}
			ctx := context.WithValue(r.Context(), spaceMemberKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SpaceFrom(ctx context.Context) (*entity.Space, bool) {
	s, ok := ctx.Value(spaceKey).(*entity.Space)
	return s, ok
}

func SpaceRoleFrom(ctx context.Context) (entity.SpaceMemberRole, bool) {
	r, ok := ctx.Value(spaceMemberKey).(entity.SpaceMemberRole)
	return r, ok
}

func writeSpaceErr(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{"code": code, "message": message},
	})
}
