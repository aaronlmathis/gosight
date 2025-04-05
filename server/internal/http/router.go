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

	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/gorilla/mux"
)

func SetupRoutes(r *mux.Router, index *store.MetricIndex, staticDir, templateDir, env string) {

	// Serve static assets
	fs := http.FileServer(http.Dir(staticDir))
	r.PathPrefix("/css/").Handler(fs)
	r.PathPrefix("/js/").Handler(fs)
	r.PathPrefix("/images/").Handler(fs)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		RenderIndexPage(w, r, templateDir, env)
	})
	r.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		RenderAgentsPage(w, r, templateDir, env)
	})
	r.HandleFunc("/containers", func(w http.ResponseWriter, r *http.Request) {
		RenderContainersPage(w, r, templateDir, env)
	})
	r.HandleFunc("/api/containers", HandleContainersAPI).Methods("GET")
	r.HandleFunc("/api/agents", HandleAgentsAPI).Methods("GET")

	meta := NewMetricMetaHandler(index)

	r.HandleFunc("/api/metrics/namespaces", meta.GetNamespaces).Methods("GET")
	r.HandleFunc("/api/metrics/subnamespaces", meta.GetSubNamespaces).Methods("GET")
	r.HandleFunc("/api/metrics/names", meta.GetMetricNames).Methods("GET")
	r.HandleFunc("/api/metrics/dimensions", meta.GetDimensions).Methods("GET")
	// ...
}
