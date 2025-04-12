/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis aaron.mathis@gmail.com

This file is part of GoSight.

GoSight is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GoSight is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GoSight. If not, see https://www.gnu.org/licenses/.
*/

// server/internal/http/router.go
// Router for HTTPServer
package httpserver

import (
	"encoding/base64"
	"net/http"
	"time"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/server/internal/usermodel"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
)

func SetupRoutes(r *mux.Router, metricIndex *store.MetricIndex, metricStore store.MetricStore,
	userStore userstore.UserStore, metaTracker *metastore.MetaTracker, authProviders map[string]gosightauth.AuthProvider, cfg *config.Config) {

	r.Handle("/",
		gosightauth.AuthMiddleware(userStore)(
			gosightauth.RequirePermission("gosight:dashboard:view",
				gosightauth.AccessLogMiddleware(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						HandleIndex(w, r, cfg, userStore)
					}),
				),
				userStore,
			),
		),
	)

	r.HandleFunc("/logout", HandleLogout).Methods("GET")
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		HandleLoginPage(w, r, authProviders)
	}).Methods("GET")

	// Start login for a provider (Google, Azure, etc.)
	r.HandleFunc("/login/start", func(w http.ResponseWriter, r *http.Request) {
		provider := r.URL.Query().Get("provider")
		if handler, ok := authProviders[provider]; ok {
			handler.StartLogin(w, r)
		} else {
			http.Error(w, "invalid provider", http.StatusBadRequest)
		}
	}).Methods("GET")

	// Handle provider callback (local or SSO)
	r.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		provider := r.URL.Query().Get("provider")

		if handler, ok := authProviders[provider]; ok {
			user, err := handler.HandleCallback(w, r)
			if err != nil {
				SetFlash(w, "Invalid username or password")
				http.Redirect(w, r, "/login", http.StatusSeeOther)

				return
			}
			//  Load roles + permissions
			user, err = userStore.GetUserWithPermissions(r.Context(), user.ID)
			if err != nil {
				utils.Error("❌ Failed to load roles for user %s: %v", user.Email, err)
				http.Error(w, "failed to load user roles", http.StatusInternalServerError)
				return
			}

			// Only enforce 2FA for local users
			if provider == "local" && user.TOTPSecret != "" {
				if gosightauth.CheckRememberMFA(r, user.ID) {
					// Valid remembered device — skip MFA
					ctx := InjectUserContext(r.Context(), user)
					createFinalSessionAndRedirect(w, r.WithContext(ctx), user)
					return
				}
				gosightauth.SavePendingMFA(user.ID, w)

				// Optional: persist 'next' (state)
				state := r.URL.Query().Get("state")
				if state != "" {
					http.SetCookie(w, &http.Cookie{
						Name:     "pending_next",
						Value:    state,
						Path:     "/",
						MaxAge:   300,
						HttpOnly: true,
						Secure:   true,
					})
				}
				http.Redirect(w, r, "/mfa", http.StatusSeeOther)
				return

			} else {

				ctx := InjectUserContext(r.Context(), user)
				utils.Debug("✅ Google user: %s", user.Email)
				utils.Debug("✅ Token will be issued with roles: %v", user.Roles)
				utils.Debug("✅ Flattened permissions: %v", gosightauth.FlattenPermissions(user.Roles))

				// create final session and redirect
				createFinalSessionAndRedirect(w, r.WithContext(ctx), user)
				return
			}

		}

		http.Error(w, "invalid provider", http.StatusBadRequest)
	}).Methods("GET", "POST")

	r.HandleFunc("/mfa", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			HandleMFAPage(w, r)
			return
		} else {

			userID, err := gosightauth.LoadPendingMFA(r)
			if err != nil || userID == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			code := r.FormValue("code")
			user, err := userStore.GetUserByID(r.Context(), userID)
			if err != nil || !gosightauth.ValidateTOTP(user.TOTPSecret, code) {
				http.Error(w, "Invalid TOTP code", http.StatusUnauthorized)
				return
			}

			// Load roles
			user, err = userStore.GetUserWithPermissions(r.Context(), user.ID)
			if err != nil {
				http.Error(w, "failed to load roles", http.StatusInternalServerError)
				return
			}

			// Set Remember me cookie
			utils.Debug("✅ MFA passed for", user.ID)

			if r.FormValue("remember") == "on" {
				utils.Debug("📦 Setting remember_mfa cookie")
				gosightauth.SetRememberMFA(w, user.ID, r)
			}

			// Inject context
			ctx := contextutil.SetUserID(r.Context(), user.ID)
			roles := gosightauth.ExtractRoleNames(user.Roles)
			ctx = contextutil.SetUserRoles(ctx, roles)
			ctx = contextutil.SetUserPermissions(ctx, gosightauth.FlattenPermissions(user.Roles))

			// Final login and redirect
			createFinalSessionAndRedirect(w, r.WithContext(ctx), user)
		}
	})

	//r.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
	//	RenderAgentsPage(w, r, cfg.Web.TemplateDir, cfg.Server.Environment)
	//})

	r.Handle("/endpoints/{endpoint_id}",
		gosightauth.AuthMiddleware(userStore)(
			gosightauth.RequirePermission("gosight:dashboard:view", // TODO Permissions
				gosightauth.AccessLogMiddleware(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						HandleEndpointDetail(w, r, cfg, metricStore, userStore, metaTracker)
					}),
				),
				userStore,
			),
		),
	)

	//r.Handle("/api/endpoints/containers", &ContainerHandler{Store: metricStore})
	//r.Handle("/api/endpoints/hosts", &HostsHandler{Store: metricStore})
	//r.HandleFunc("/api/agents", HandleAgentsAPI).Methods("GET")

	meta := NewMetricMetaHandler(metricIndex, metricStore)
	r.HandleFunc("/api", meta.GetNamespaces).Methods("GET")
	r.HandleFunc("/api/{namespace}/{sub}/{metric}/latest", meta.GetLatestValue).Methods("GET")
	r.HandleFunc("/api/{namespace}/{sub}/{metric}/data", meta.GetMetricData).Methods("GET")
	r.HandleFunc("/api/{namespace}/{sub}/dimensions", meta.GetDimensions).Methods("GET")
	r.HandleFunc("/api/{namespace}/{sub}", meta.GetMetricNames).Methods("GET")
	r.HandleFunc("/api/{namespace}", meta.GetSubNamespaces).Methods("GET")
	r.HandleFunc("/api/query", meta.HandleAPIQuery).Methods("GET")

	// ...
}

func createFinalSessionAndRedirect(w http.ResponseWriter, r *http.Request, user *usermodel.User) {
	traceID, _ := contextutil.GetTraceID(r.Context())
	roles := gosightauth.ExtractRoleNames(user.Roles)

	token, err := gosightauth.GenerateToken(user.ID, roles, traceID)
	if err != nil {
		utils.Error("❌ Failed to generate session token: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	gosightauth.SetSessionCookie(w, token)

	// Clear any pending MFA/session leftovers
	gosightauth.ClearCookie(w, "pending_mfa")
	http.SetCookie(w, &http.Cookie{
		Name:     "pending_next",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	// handle redirect after login
	next := "/"
	if state := r.URL.Query().Get("state"); state != "" {
		if decoded, err := base64.URLEncoding.DecodeString(state); err == nil {
			next = string(decoded)
		}
	}
	if c, err := r.Cookie("pending_next"); err == nil {
		next = c.Value
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}
