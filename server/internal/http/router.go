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
	"net/http"
	"path/filepath"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
)

func (s *HttpServer) setupRoutes() {
	s.setupStaticRoutes()
	s.setupAuthRoutes()
	s.setupMetricExplorerRoutes()
	s.setupActivityRoutes()
	s.setupEndpointRoutes()
	s.setupAPIRoutes()
	s.setupIndexRoutes()
	s.setupWebSocketRoutes()
}

func (s *HttpServer) setupStaticRoutes() {
	staticFS := http.FileServer(http.Dir(s.Sys.Cfg.Web.StaticDir))

	/*
		cacheWrapper := func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", "public, max-age=86400")
				h.ServeHTTP(w, r)
			})
		}
	*/

	// Serve static assets like /js/, /css/, /images/ directly from StaticDir
	// For Production ----
	//s.Router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", staticFS))
	//s.Router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", staticFS))

	// For local dev
	serveWithMime := func(prefix string, subdir string, contentTypeMap map[string]string) http.Handler {
		return http.StripPrefix(prefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ext := filepath.Ext(r.URL.Path)
			if ct, ok := contentTypeMap[ext]; ok {
				w.Header().Set("Content-Type", ct)
			}
			fullPath := filepath.Join(s.Sys.Cfg.Web.StaticDir, subdir, filepath.Base(r.URL.Path))
			http.ServeFile(w, r, fullPath)
		}))
	}

	// Register /css/ and /js/
	s.Router.PathPrefix("/css/").Handler(serveWithMime("/css/", "css", map[string]string{
		".css": "text/css",
	}))

	s.Router.PathPrefix("/js/").Handler(serveWithMime("/js/", "js", map[string]string{
		".js": "application/javascript",
	}))

	s.Router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", staticFS))
}

// setupAuthRoutes sets up the authentication routes for the HTTP server.
// It includes routes for login, logout, MFA, and the callback from the auth provider.
// The routes are protected by middleware that checks for injects context and trace identifiers.
// The routes are also logged for access control.
func (s *HttpServer) setupAuthRoutes() {
	withAccessLog := gosightauth.AccessLogMiddleware
	withAuth := gosightauth.AuthMiddleware(s.Sys.Stores.Users)

	// These routes require no auth — public-facing
	s.Router.Handle("/login", withAccessLog(http.HandlerFunc(s.HandleLogin))).Methods("GET")
	s.Router.Handle("/login/start", withAccessLog(http.HandlerFunc(s.HandleLoginStart))).Methods("GET")
	s.Router.Handle("/callback", withAccessLog(http.HandlerFunc(s.HandleCallback))).Methods("GET", "POST")
	s.Router.Handle("/mfa", withAccessLog(http.HandlerFunc(s.HandleMFA))).Methods("GET", "POST")

	s.Router.Handle("/logout",
		withAccessLog(
			withAuth(
				http.HandlerFunc(s.HandleLogout),
			),
		),
	).Methods("GET")
}

// setupIndexRoutes sets up the index routes for the HTTP server.
// It includes routes for the index page and the dashboard page.
// The routes are protected by middleware that checks for injects context and trace identifiers.
// The routes are also logged for access control.
func (s *HttpServer) setupIndexRoutes() {
	s.Router.Handle("/",
		gosightauth.AuthMiddleware(s.Sys.Stores.Users)(
			gosightauth.RequirePermission("gosight:dashboard:view",
				gosightauth.AccessLogMiddleware(
					http.HandlerFunc(s.HandleIndexPage),
				),
				s.Sys.Stores.Users,
			),
		),
	)
}

// setupMetricExplorerRoutes sets up the metric explorer routes for the HTTP server.
// It includes routes for viewing the metric explorer page and the metric detail page.
// The routes are protected by middleware that checks for injects context and trace identifiers.
// The routes are also logged for access control.
func (s *HttpServer) setupMetricExplorerRoutes() {
	s.Router.Handle("/metrics",
		gosightauth.AuthMiddleware(s.Sys.Stores.Users)(
			gosightauth.RequirePermission("gosight:dashboard:view",
				gosightauth.AccessLogMiddleware(
					http.HandlerFunc(s.HandleMetricExplorerPage),
				),
				s.Sys.Stores.Users,
			),
		),
	)
}

// setupActivityRoutes sets up the activity routes for the HTTP server.
// It includes routes for viewing and managing activity logs.
// The routes are protected by middleware that checks for injects context and trace identifiers.
// The routes are also logged for access control.

func (s *HttpServer) setupActivityRoutes() {
	s.Router.Handle("/activity",
		gosightauth.AuthMiddleware(s.Sys.Stores.Users)(
			gosightauth.RequirePermission("gosight:dashboard:view",
				gosightauth.AccessLogMiddleware(
					http.HandlerFunc(s.HandleActivityPage),
				),
				s.Sys.Stores.Users,
			),
		),
	)
	s.Router.Handle("/activity/{stream}",
		gosightauth.AuthMiddleware(s.Sys.Stores.Users)(
			gosightauth.RequirePermission("gosight:dashboard:view",
				gosightauth.AccessLogMiddleware(
					http.HandlerFunc(s.HandleEndpointDetail),
				),
				s.Sys.Stores.Users,
			),
		),
	)
}

// setupEndpointRoutes sets up the endpoint routes for the HTTP server.
// It includes routes for fetching the endpoint page and the endpoint detail page.
// The routes are protected by middleware that checks for injects context and trace identifiers.
// The routes are also logged for access control.

func (s *HttpServer) setupEndpointRoutes() {
	s.Router.Handle("/endpoints",
		gosightauth.AuthMiddleware(s.Sys.Stores.Users)(
			gosightauth.RequirePermission("gosight:dashboard:view",
				gosightauth.AccessLogMiddleware(
					http.HandlerFunc(s.HandleEndpointPage),
				),
				s.Sys.Stores.Users,
			),
		),
	)
	s.Router.Handle("/endpoints/{endpoint_id}",
		gosightauth.AuthMiddleware(s.Sys.Stores.Users)(
			gosightauth.RequirePermission("gosight:dashboard:view",
				gosightauth.AccessLogMiddleware(
					http.HandlerFunc(s.HandleEndpointDetail),
				),
				s.Sys.Stores.Users,
			),
		),
	)

}

// setupAPIRoutes sets up the API routes for the HTTP server.
// It includes routes for fetching namespaces, sub-namespaces, metric names, dimensions,
// and metric data.
func (s *HttpServer) setupAPIRoutes() {
	api := s.Router.PathPrefix("/api/v1").Subrouter()

	withAuth := gosightauth.AuthMiddleware(s.Sys.Stores.Users)

	secure := func(permission string, handler http.HandlerFunc) http.Handler {
		return withAuth(gosightauth.RequirePermission(permission, handler, s.Sys.Stores.Users))
	}
	// Endpoint APIs
	api.Handle("/endpoints", secure("gosight:api:endpoints:view", http.HandlerFunc(s.HandleEndpointsAPI))).Methods("GET")
	api.Handle("/endpoints/{endpointType}", secure("gosight:api:endpoints:view", http.HandlerFunc(s.HandleEndpointsByTypeAPI))).Methods("GET")

	// Logs
	api.Handle("/logs", secure("gosight:api:logs:view", http.HandlerFunc(s.HandleLogAPI))).Methods("GET")
	api.Handle("/logs/latest", secure("gosight:api:logs:view", http.HandlerFunc(s.HandleRecentLogs))).Methods("GET")

	// Events and Alerts
	api.Handle("/events", secure("gosight:api:events:view", http.HandlerFunc(s.HandleEventsAPI))).Methods("GET")

	api.Handle("/alerts", secure("gosight:api:events:view", http.HandlerFunc(s.HandleAlertsAPI))).Methods("GET")

	// Metrics and queries
	api.Handle("/query", secure("gosight:api:metrics:query", http.HandlerFunc(s.HandleAPIQuery))).Methods("GET")
	api.Handle("/exportquery", secure("gosight:api:metrics:export", http.HandlerFunc(s.HandleExportQuery))).Methods("GET")

	// Metadata discovery endpoints
	api.Handle("/", secure("gosight:api:metrics:meta", http.HandlerFunc(s.GetNamespaces))).Methods("GET")
	api.Handle("/{namespace}/{sub}/{metric}/dimensions", secure("gosight:api:metrics:meta", http.HandlerFunc(s.GetMetricDimensions))).Methods("GET")
	api.Handle("/{namespace}/{sub}/{metric}/data", secure("gosight:api:metrics:read", http.HandlerFunc(s.GetMetricData))).Methods("GET")
	api.Handle("/{namespace}/{sub}/{metric}/latest", secure("gosight:api:metrics:read", http.HandlerFunc(s.GetMetricLatest))).Methods("GET")

	api.Handle("/{namespace}/{sub}", secure("gosight:api:metrics:meta", http.HandlerFunc(s.GetMetricNames))).Methods("GET")
	api.Handle("/{namespace}", secure("gosight:api:metrics:meta", http.HandlerFunc(s.GetSubNamespaces))).Methods("GET")

}

func (s *HttpServer) setupWebSocketRoutes() {
	s.Router.HandleFunc("/ws/metrics", s.Sys.WSHub.Metrics.ServeWS)
	s.Router.HandleFunc("/ws/alerts", s.Sys.WSHub.Alerts.ServeWS)
	s.Router.HandleFunc("/ws/events", s.Sys.WSHub.Events.ServeWS)
	s.Router.HandleFunc("/ws/logs", s.Sys.WSHub.Logs.ServeWS)
}
