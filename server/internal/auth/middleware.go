// session.go: Handles signed cookies or token generation/validation.
package gosightauth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/server/internal/usermodel"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func RequirePermission(required string, next http.Handler, userStore userstore.UserStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Permissions missing from context?
		perms, ok := contextutil.GetUserPermissions(ctx)
		if !ok || len(perms) == 0 {
			// Try to fetch fresh perms from DB
			userID, ok := contextutil.GetUserID(ctx)
			if ok {
				user, err := userStore.GetUserWithPermissions(ctx, userID)
				if err == nil {
					perms = FlattenPermissions(user.Roles)
					ctx = contextutil.SetUserPermissions(ctx, perms)
					r = r.WithContext(ctx) // update request context
				}
			} else {
				// User is not authenticated
				// Set forbidden flag in context
				ctx = contextutil.SetForbidden(ctx)
				r = r.WithContext(ctx)

			}
		}

		if !HasPermission(ctx, required) {
			//utils.Debug("RequirePermission: missing %s", required)

			if isAPIRequest(r) {
				http.Error(w, "forbidden", http.StatusForbidden)
			} else {
				http.Redirect(w, r, "/unauthorized", http.StatusSeeOther)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequireAnyPermissionWithStore(store userstore.UserStore, required ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			perms, ok := contextutil.GetUserPermissions(ctx)

			if !ok || len(perms) == 0 {
				if userID, ok := contextutil.GetUserID(ctx); ok {
					user, err := store.GetUserWithPermissions(ctx, userID)
					if err == nil {
						perms = FlattenPermissions(user.Roles)
						ctx = contextutil.SetUserPermissions(ctx, perms)
						r = r.WithContext(ctx)
					}
				}
			}

			if !HasAnyPermission(ctx, required...) {
				//utils.Debug("RequireAnyPermission: missing one of %v", required)

				if isAPIRequest(r) {
					http.Error(w, "forbidden", http.StatusForbidden)
				} else {
					http.Redirect(w, r, "/unauthorized", http.StatusSeeOther)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func AuthMiddleware(userStore userstore.UserStore) mux.MiddlewareFunc {
	//utils.Debug("AuthMiddleware: Injecting user context")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := GetSessionToken(r)
			if err != nil {
				//utils.Debug("AuthMiddleWare: GetSessionToken failed: %v", err)
				if isAPIRequest(r) {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
				} else {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
				}
				return
			}
			//utils.Debug("AuthMiddleware: Token found: %s", token)
			claims, err := ValidateToken(token)
			if err != nil {
				//utils.Debug("AuthMiddleWare: ValidateToken failed: %v", err)
				if isAPIRequest(r) {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
				} else {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
				}
				return
			}
			//utils.Debug("AuthMiddleware: Token claims: %v", claims)
			userID := claims.UserID
			traceID := claims.TraceID

			var user *usermodel.User
			var roleNames []string
			var perms []string
			rolesTTL := 10 * time.Minute
			refreshedAt := time.Unix(claims.RolesRefreshedAt, 0)

			// Revalidate if roles are stale or missing
			if len(claims.Roles) == 0 || time.Since(refreshedAt) > rolesTTL {
				user, err = userStore.GetUserWithPermissions(r.Context(), userID)
				if err != nil {
					//utils.Debug("Failed to reload user roles: %v", err)
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				roleNames = ExtractRoleNames(user.Roles)
				perms = FlattenPermissions(user.Roles)
				//utils.Debug("Flattened permissions from DB: %v", perms)
			} else {
				// Roles are fresh, use from token
				roleNames = claims.Roles

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

			if user != nil && len(user.Roles) > 0 {
				perms = FlattenPermissions(user.Roles)
				ctx = contextutil.SetUserPermissions(ctx, perms)
				//utils.Debug("Revalidated user: %s, permissions: %v", user.ID, perms)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Use existing X-Trace-ID or generate one
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.NewString()
		}

		// Set trace ID into context and response header
		ctx := contextutil.SetTraceID(r.Context(), traceID)
		w.Header().Set("X-Trace-ID", traceID)

		// Wrap writer to capture status code
		rr := &statusRecorder{ResponseWriter: w, status: 200}

		// ⏱Call next handler
		next.ServeHTTP(rr, r.WithContext(ctx))

		duration := time.Since(start)

		// Structured fields
		userID, _ := contextutil.GetUserID(ctx)
		roles, _ := contextutil.GetUserRoles(ctx)
		perms, _ := contextutil.GetUserPermissions(ctx)

		if userID == "" {
			userID = "anonymous"
		}

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

// isAPIRequest checks if the request is for an API endpoint.
func isAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/")
}
