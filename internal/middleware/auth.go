package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go-template/internal/jwtauth"
)

type contextKey string

const claimsKey contextKey = "claims"

func Auth(jm jwtauth.Tokenizer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeErr(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			claims, err := jm.Parse(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				writeErr(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(claimsKey).(*jwtauth.Claims)
			if !ok || claims.Role != role {
				writeErr(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ClaimsFrom(ctx context.Context) (*jwtauth.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*jwtauth.Claims)
	return claims, ok
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
