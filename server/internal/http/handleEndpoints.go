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
	"net/http"

	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/http/templates"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
)

type HostRow struct {
	Agent   model.Agent
	Metrics map[string]string // uptime, cpu %, mem %, etc.
}

type ContainerRow struct {
	ID     string
	Name   string
	Image  string
	Status string
	CPU    string
	Mem    string
	RX     string
	TX     string
	Uptime string
}

func (s *HttpServer) HandleEndpointPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if forbidden, ok := ctx.Value("forbidden").(bool); ok && forbidden {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, ok := contextutil.GetUserID(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := s.UserStore.GetUserWithPermissions(ctx, userID)
	if err != nil {
		utils.Error("Failed to load user %s: %v", userID, err)
		http.Error(w, "failed to load user", http.StatusInternalServerError)
		return
	}
	pageData := templates.TemplateData{
		Title: "Endpoints",
		User:  user,
		Breadcrumbs: []templates.Breadcrumb{
			{Label: "Endpoints"},
		},
	}

	err = templates.RenderTemplate(w, "dashboard/layout_endpoints", pageData)

	if err != nil {
		http.Error(w, "template error", 500)
	}
}

func (s *HttpServer) HandleEndpointDetail(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	endpointID := vars["endpoint_id"]
	ctx := r.Context()
	utils.Debug("Endpoint ID: %s", endpointID)

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
		utils.Error("Failed to load user %s: %v", userID, err)
		http.Error(w, "failed to load user", http.StatusInternalServerError)
		return
	}

	meta, _ := s.MetaTracker.Get(endpointID)

	// Build Template data based on endpoint_id
	pageData := templates.TemplateData{
		Title:  "Endpoints",
		User:   user,
		Labels: map[string]string{"endpoint_id": endpointID},
		Breadcrumbs: []templates.Breadcrumb{
			{Label: "Endpoints", URL: "/endpoints"},
			{Label: endpointID},
		},
		Meta: meta,
	}

	err = templates.RenderTemplate(w, "dashboard/layout_main", pageData)
	if err != nil {
		utils.Error("Template error: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}
