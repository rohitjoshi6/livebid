package middleware

import (
	"net/http"
	"strings"

	"github.com/rohitjoshi6/livebid/backend/internal/auth"
	"github.com/rohitjoshi6/livebid/backend/internal/httpx"
)

func RequireAuth(tokens *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "missing_token", "Bearer token is required.")
				return
			}
			tokenValue, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || tokenValue == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "invalid_token", "Authorization header must use Bearer token format.")
				return
			}
			claims, err := tokens.Parse(tokenValue)
			if err != nil {
				httpx.WriteError(w, http.StatusUnauthorized, "invalid_token", "Token is invalid or expired.")
				return
			}
			next.ServeHTTP(w, r.WithContext(auth.ContextWithClaims(r.Context(), claims)))
		})
	}
}
