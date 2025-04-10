// session.go: Handles signed cookies or token generation/validation.
package gosightauth

import (
	"net/http"
	"net/url"

	"github.com/aaronlmathis/gosight/server/internal/contextutil"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := GetSessionUserID(r)
		if err != nil {
			// Redirect to login with next=originalPath
			http.Redirect(w, r, "/login?next="+url.QueryEscape(r.URL.RequestURI()), http.StatusSeeOther)
			return
		}

		ctx := contextutil.SetUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
