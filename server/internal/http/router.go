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

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
)

func SetupRoutes(r *mux.Router, metricIndex *store.MetricIndex, metricStore store.MetricStore, userStore userstore.UserStore, authProviders map[string]gosightauth.AuthProvider, cfg *config.Config) {

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
		HandleLoginPage(w, r, authProviders, cfg.Web.TemplateDir)
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
				utils.Debug("❌ Login failed for provider %s: %v", provider, err)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			//  Load roles + permissions
			user, err = userStore.GetUserWithPermissions(r.Context(), user.ID)
			if err != nil {
				utils.Error("❌ Failed to load roles for user %s: %v", user.Email, err)
				http.Error(w, "failed to load user roles", http.StatusInternalServerError)
				return
			}

			//  Inject context (for immediate handlers, or log/debug)
			ctx := contextutil.SetUserID(r.Context(), user.ID)
			userRoles := gosightauth.ExtractRoleNames(user.Roles)
			ctx = contextutil.SetUserRoles(ctx, userRoles)
			ctx = contextutil.SetUserPermissions(ctx, gosightauth.FlattenPermissions(user.Roles))

			//  Not passed to next here, but could be for inline chaining
			traceID, _ := contextutil.GetTraceID(r.Context())
			//  Set session + redirect
			token, err := gosightauth.GenerateToken(user.ID, userRoles, traceID)
			if err != nil {
				utils.Error("❌ Failed to generate session token for user %s: %v", user.Email, err)
				http.Error(w, "server error", http.StatusInternalServerError)
				return
			}
			gosightauth.SetSessionCookie(w, token)

			state := r.URL.Query().Get("state")
			var next string

			if state != "" {
				decoded, err := base64.URLEncoding.DecodeString(state)
				if err == nil {
					next = string(decoded)
				}
			}
			if next == "" {
				next = "/"
			}
			utils.Debug("✅ Google user: %s", user.Email)
			utils.Debug("✅ Token will be issued with roles: %v", userRoles)
			utils.Debug("✅ Flattened permissions: %v", gosightauth.FlattenPermissions(user.Roles))
			http.Redirect(w, r, next, http.StatusSeeOther)
			return
		}

		http.Error(w, "invalid provider", http.StatusBadRequest)
	}).Methods("GET", "POST")

	r.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		RenderAgentsPage(w, r, cfg.Web.TemplateDir, cfg.Server.Environment)
	})

	r.Handle("/endpoints/{endpoint_id}",
		gosightauth.AuthMiddleware(userStore)(
			gosightauth.RequirePermission("gosight:dashboard:view", // TODO Permissions
				gosightauth.AccessLogMiddleware(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						HandleEndpointDetail(w, r, cfg, metricStore, userStore)
					}),
				),
				userStore,
			),
		),
	)

	r.HandleFunc("/mockup", func(w http.ResponseWriter, r *http.Request) {
		RenderMockupPage(w, r, cfg.Web.TemplateDir)
	})
	r.Handle("/api/endpoints/containers", &ContainerHandler{Store: metricStore})
	r.Handle("/api/endpoints/hosts", &HostsHandler{Store: metricStore})
	r.HandleFunc("/api/agents", HandleAgentsAPI).Methods("GET")

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
