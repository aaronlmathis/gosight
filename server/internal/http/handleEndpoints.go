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

// File: server/internal/http/handleEndpoint.go
// Description: This file contains the handlers for the endpoint details page and API.
// It includes the main endpoint details page and the API handler for fetching endpoint metrics.

package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/http/templates"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
)

func (s *HttpServer) HandleEndpointDetail(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	endpointID := vars["endpoint_id"]
	ctx := r.Context()

	// Check for forbidden access first
	if forbidden, ok := ctx.Value("forbidden").(bool); ok && forbidden {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if user is authenticated
	userID, ok := contextutil.GetUserID(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// Check if user has permission to view the dashboard
	user, err := s.UserStore.GetUserWithPermissions(ctx, userID)
	if err != nil {
		utils.Error("❌ Failed to load user %s: %v", userID, err)
		http.Error(w, "failed to load user", http.StatusInternalServerError)
		return
	}

	// Build Template data based on endpoint_id
	data, err := templates.BuildHostDashboardData(ctx, s.MetricStore, s.MetaTracker, user, endpointID)
	if err != nil {
		utils.Debug("failed to build host dashboard data: %v", err)
	}
	fmt.Printf("🧠 Template Meta: %+v\n", data.Meta)
	// Set breadcrumbs and endpoint id
	data.Title = "Host: " + endpointID
	data.Labels["EndpointID"] = endpointID

	err = templates.RenderTemplate(w, "dashboard/layout_main", data)
	if err != nil {
		utils.Error("❌ Template error: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func (s *HttpServer) EndpointDetailsAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpointID := vars["endpoint_id"]

	// Basic labels
	labels := map[string]string{
		"endpoint_id": endpointID,
	}

	instantNames := templates.GetMetricNames(templates.HostMetrics, true)
	fmt.Println("🧪 Querying instant metrics:", instantNames)

	rows, err := s.MetricStore.QueryMultiInstant(instantNames, labels)
	if err != nil {
		http.Error(w, "error fetching metrics", http.StatusInternalServerError)
		return
	}

	// Prepare structured response
	response := map[string]interface{}{
		"endpoint_id": endpointID,
		"metrics":     map[string]float64{},
		"labels":      map[string]string{},
		"timestamp":   time.Now().UnixMilli(),
	}

	// Populate response metrics
	for _, row := range rows {
		metricName := row.Tags["__name__"]
		response["metrics"].(map[string]float64)[metricName] = row.Value
	}

	// Add labels or extra info if needed
	if endpoint, ok := s.MetaTracker.Get(endpointID); ok {

		response["labels"].(map[string]string)["os"] = endpoint.OS

	}

	// Set JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
