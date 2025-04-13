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
	s.setupEndpointRoutes()
	s.setupAPIRoutes()
	s.setupIndexRoutes()
	s.setupWebSocketRoutes()
}

func (s *HttpServer) setupStaticRoutes() {
	staticFS := http.FileServer(http.Dir(s.Config.Web.StaticDir))

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
			fullPath := filepath.Join(s.Config.Web.StaticDir, subdir, filepath.Base(r.URL.Path))
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

func (s *HttpServer) setupAuthRoutes() {
	s.Router.HandleFunc("/callback", s.HandleCallback).Methods("GET", "POST")
	s.Router.HandleFunc("/login/start", s.HandleLoginStart).Methods("GET")
	s.Router.HandleFunc("/logout", s.HandleLogout).Methods("GET")
	s.Router.HandleFunc("/login", s.HandleLogin).Methods("GET")
	s.Router.HandleFunc("/mfa", s.HandleMFA).Methods("GET", "POST")
}

func (s *HttpServer) setupIndexRoutes() {
	s.Router.HandleFunc("/", s.HandleIndex).Methods("GET")
}

func (s *HttpServer) setupEndpointRoutes() {
	s.Router.Handle("/endpoints/{endpoint_id}",
		gosightauth.AuthMiddleware(s.UserStore)(
			gosightauth.RequirePermission("gosight:dashboard:view",
				gosightauth.AccessLogMiddleware(
					http.HandlerFunc(s.HandleEndpointDetail),
				),
				s.UserStore,
			),
		),
	)
}

// setupAPIRoutes sets up the API routes for the HTTP server.
// It includes routes for fetching namespaces, sub-namespaces, metric names, dimensions,
// and metric data.
func (s *HttpServer) setupAPIRoutes() {

	api := s.Router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/query", s.HandleAPIQuery).Methods("GET")
	api.HandleFunc("/endpoints/{endpoint_id}", s.EndpointDetailsAPIHandler).Methods("GET")
	api.HandleFunc("/", s.GetNamespaces).Methods("GET")
	api.HandleFunc("/{namespace}/{sub}/{metric}/latest", s.GetLatestValue).Methods("GET")
	api.HandleFunc("/{namespace}/{sub}/{metric}/data", s.GetMetricData).Methods("GET")
	api.HandleFunc("/{namespace}/{sub}/dimensions", s.GetDimensions).Methods("GET")
	api.HandleFunc("/{namespace}/{sub}", s.GetMetricNames).Methods("GET")
	api.HandleFunc("/{namespace}", s.GetSubNamespaces).Methods("GET")

}

func (s *HttpServer) setupWebSocketRoutes() {
	s.Router.HandleFunc("/ws/metrics", s.WebSocket.ServeWS)
}
