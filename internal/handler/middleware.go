package handler

import (
	"context"
	"net/http"
	"strings"

	"labago/internal/jwt"
)

type contextKey string

const UserIDKey contextKey = "userID"

// AuthMiddleware проверяет JWT из заголовка Authorization и кладёт userID в контекст запроса.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}
		claims, err := jwt.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next(w, r.WithContext(ctx))
	}
}
