package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"nerion/internal/domain"
	"nerion/internal/entity"
)

type apiKeyContextKey string

const apiKeyCtxKey apiKeyContextKey = "api_key"

// APIKeyAuth extracts X-Api-Key header, validates it against the DB, and stores the key in context.
// Asynchronously updates last_used_at and logs an audit entry to avoid adding latency.
func APIKeyAuth(repo domain.APIKeyRepository, auditRepo domain.AuditRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawKey := r.Header.Get("X-Api-Key")
			if rawKey == "" {
				writeAPIKeyErr(w, http.StatusUnauthorized, "unauthorized", "Требуется API ключ")
				return
			}
			h := sha256.Sum256([]byte(rawKey))
			hash := hex.EncodeToString(h[:])
			key, err := repo.FindByHash(r.Context(), hash)
			if err != nil {
				writeAPIKeyErr(w, http.StatusUnauthorized, "unauthorized", "Недействительный API ключ")
				return
			}
			go func() {
				_ = repo.UpdateLastUsed(context.Background(), key.ID)
				_ = auditRepo.Log(context.Background(), &entity.AuditEntry{
					SpaceID:    &key.SpaceID,
					Action:     "api_key.request",
					EntityType: "api_key",
					EntityID:   key.KeyPrefix,
					Meta:       map[string]any{"method": r.Method, "path": r.URL.Path},
				})
			}()
			ctx := context.WithValue(r.Context(), apiKeyCtxKey, key)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func APIKeyFrom(ctx context.Context) (*entity.APIKey, bool) {
	k, ok := ctx.Value(apiKeyCtxKey).(*entity.APIKey)
	return k, ok
}

func writeAPIKeyErr(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}
