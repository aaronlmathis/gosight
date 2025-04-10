// session.go: Handles signed cookies or token generation/validation.
package gosightauth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/server/internal/usermodel"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func AuthMiddleware(userStore userstore.UserStore) mux.MiddlewareFunc {
	utils.Debug("🔑 AuthMiddleware: Injecting user context")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := GetSessionToken(r)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := ValidateToken(token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID := claims.UserID
			traceID := claims.TraceID

			var user *usermodel.User
			var roleNames []string
			rolesTTL := 10 * time.Minute
			refreshedAt := time.Unix(claims.RolesRefreshedAt, 0)

			// Revalidate if roles are stale or missing
			if len(claims.Roles) == 0 || time.Since(refreshedAt) > rolesTTL {
				user, err = userStore.GetUserWithPermissions(r.Context(), userID)
				if err != nil {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				roleNames = ExtractRoleNames(user.Roles)
			} else {
				// Roles are fresh, use from token
				roleNames = claims.Roles
				user = &usermodel.User{ID: userID} // minimal fallback user
			}

			// Inject trace ID
			ctx := r.Context()
			if traceID != "" {
				ctx = contextutil.SetTraceID(ctx, traceID)
			} else if _, ok := contextutil.GetTraceID(ctx); !ok {
				ctx = contextutil.SetTraceID(ctx, uuid.New().String())
			}

			ctx = contextutil.SetUserID(ctx, userID)
			ctx = contextutil.SetUserRoles(ctx, roleNames)

			// Only flatten perms if we loaded real roles from DB
			if user != nil && len(user.Roles) > 0 {
				ctx = contextutil.SetUserPermissions(ctx, FlattenPermissions(user.Roles))
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 🔍 Use existing X-Trace-ID or generate one
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.NewString()
		}

		// 🔁 Set trace ID into context and response header
		ctx := contextutil.SetTraceID(r.Context(), traceID)
		w.Header().Set("X-Trace-ID", traceID)

		// 🧱 Wrap writer to capture status code
		rr := &statusRecorder{ResponseWriter: w, status: 200}

		// ⏱️ Call next handler
		next.ServeHTTP(rr, r.WithContext(ctx))

		duration := time.Since(start)

		// 📦 Structured fields
		userID, _ := contextutil.GetUserID(ctx)
		roles, _ := contextutil.GetUserRoles(ctx)
		perms, _ := contextutil.GetUserPermissions(ctx)

		entry := map[string]interface{}{
			"timestamp":   time.Now().Format(time.RFC3339),
			"trace_id":    traceID,
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      rr.status,
			"duration_ms": duration.Milliseconds(),
			"user_id":     userID,
			"roles":       roles,
			"permissions": perms,
			"user_agent":  r.UserAgent(),
			"ip":          r.RemoteAddr,
		}

		logJSON, _ := json.Marshal(entry)
		utils.Access(string(logJSON)) // or send to file/syslog/etc
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
