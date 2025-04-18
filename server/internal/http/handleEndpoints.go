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
	"fmt"
	"net/http"
	"strings"

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

	agents := s.AgentTracker.GetAgents()
	utils.Debug("Agents: %v", agents)
	active := map[string]bool{}
	var hosts []HostRow

	for _, agent := range agents {
		if strings.HasPrefix(agent.EndpointID, "ctr-") {
			continue
		}
		active[agent.AgentID] = true
		agent.EndpointID = agent.Labels["endpoint_id"] // TODO figure out why Endpoint_ID is empty

		hostMetrics, _ := s.MetricStore.QueryMultiInstant([]string{
			"system.host.uptime", "system.host.procs", "system.mem.free",
			"system.mem.used_percent", "system.cpu.percent",
			"system.host.users_loggedin", "system.host.info",
		}, map[string]string{"hostname": agent.Hostname})

		hostMap := make(map[string]string)
		for _, row := range hostMetrics {
			switch row.Tags["__name__"] {
			case "system.host.uptime":
				hostMap["uptime"] = templates.FormatUptime(row.Value)
			case "system.mem.free":
				hostMap["mem_free"] = templates.HumanizeBytes(row.Value)
			case "system.mem.used_percent":
				hostMap["mem"] = fmt.Sprintf("%.1f%%", row.Value)
			case "system.cpu.percent":
				hostMap["cpu"] = fmt.Sprintf("%.1f%%", row.Value)
			case "system.host.procs":
				hostMap["procs"] = fmt.Sprintf("%.0f", row.Value)
			case "system.host.users_loggedin":
				hostMap["users"] = fmt.Sprintf("%.0f", row.Value)
			case "system.host.info":
				hostMap["arch"] = row.Tags["architecture"]
				hostMap["os"] = row.Tags["os"]
				hostMap["platform"] = fmt.Sprintf("%s %s", row.Tags["platform"], row.Tags["platform_version"])
				hostMap["version"] = row.Tags["version"]
			}
		}
		utils.Debug("HostMap: %v", hostMap)
		hosts = append(hosts, HostRow{
			Agent:   agent,
			Metrics: hostMap,
		})
	}

	err = templates.RenderTemplate(w, "dashboard/layout_endpoints", map[string]any{
		"Title": "Endpoints",
		"User":  user,
		"Hosts": hosts,
	})
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

	// Build Template data based on endpoint_id
	data, err := templates.BuildHostDashboardData(ctx, s.MetricStore, s.MetaTracker, user, endpointID)
	if err != nil {
		utils.Debug("failed to build host dashboard data: %v", err)
	}
	fmt.Printf("Template Meta: %+v\n", data.Meta)
	// Set breadcrumbs and endpoint id
	data.Title = "Host: " + endpointID
	data.Labels["EndpointID"] = endpointID

	err = templates.RenderTemplate(w, "dashboard/layout_main", data)
	if err != nil {
		utils.Error("Template error: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}
