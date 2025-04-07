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

func SetupRoutes(r *mux.Router, metricIndex *store.MetricIndex, metricStore store.MetricStore, staticDir, templateDir, env string) {

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		HandleIndex(w, r, templateDir, env)
	})
	r.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		RenderAgentsPage(w, r, templateDir, env)
	})
	r.HandleFunc("/endpoints", func(w http.ResponseWriter, r *http.Request) {
		HandleEndpoints(w, r, templateDir)
	})
	r.HandleFunc("/mockup", func(w http.ResponseWriter, r *http.Request) {
		RenderMockupPage(w, r, templateDir)
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
