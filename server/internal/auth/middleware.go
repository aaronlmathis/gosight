// session.go: Handles signed cookies or token generation/validation.
package gosightauth

import (
	"context"
	"net/http"
)

type ctxKey string

const userIDKey ctxKey = "user_id"

// Middleware to require auth
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := GetSessionUserID(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), userIDKey, userID))
		next.ServeHTTP(w, r)
	})
}
