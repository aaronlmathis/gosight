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
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/http/templates"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
)

// Endpoints Page Handler

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

	user, err := s.Sys.Stores.Users.GetUserWithPermissions(ctx, userID)
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

	err = templates.RenderTemplate(w, "layout_dashboard", "dashboard_endpoints", pageData)

	if err != nil {
		http.Error(w, "template error", 500)
	}
}

// HandleEndpointDetail serves the endpoint detail page.

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
	user, err := s.Sys.Stores.Users.GetUserWithPermissions(ctx, userID)
	if err != nil {
		utils.Error("Failed to load user %s: %v", userID, err)
		http.Error(w, "failed to load user", http.StatusInternalServerError)
		return
	}

	meta, _ := s.Sys.Tele.Meta.Get(endpointID)

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

	err = templates.RenderTemplate(w, "layout_dashboard", "dashboard_endpoint_detail", pageData)
	if err != nil {
		utils.Error("Template error: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

// Endpoints API Handler

type EndpointFilter struct {
	EndpointID string
	Hostname   string
	Status     string
	HostID     string
	IP         string // ← Add this
	OS         string
	Arch       string

	Tags        map[string]string
	LastSeenMin time.Duration
	LastSeenMax time.Duration
	Limit       int
	Sort        string
	Order       string
}

func ParseEndpointFilters(r *http.Request) EndpointFilter {
	q := r.URL.Query()
	tags := utils.ParseTagString(q.Get("tags"))

	parseDur := func(val string) time.Duration {
		d, _ := time.ParseDuration(val)
		return d
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 1000
	}

	return EndpointFilter{
		EndpointID:  q.Get("endpointID"),
		Hostname:    q.Get("hostname"),
		Status:      q.Get("status"),
		HostID:      q.Get("hostID"),
		IP:          q.Get("ip"),
		OS:          q.Get("os"),
		Arch:        q.Get("arch"),
		Tags:        tags,
		LastSeenMin: parseDur(q.Get("lastSeenMin")),
		LastSeenMax: parseDur(q.Get("lastSeenMax")),
		Limit:       limit,
		Sort:        q.Get("sort"),
		Order:       strings.ToLower(q.Get("order")),
	}
}

// HandleEndpointsAPI returns a JSON list of active endpoints
// It supports querying or listing all endpoints as well as /api/v1/endpoints/{endpoint_type} (hosts / containers)
func (s *HttpServer) HandleEndpointsAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := ParseEndpointFilters(r)

	agents, err := s.Sys.Stores.Data.ListAgents(ctx)
	if err != nil {
		http.Error(w, "failed to load agents", http.StatusInternalServerError)
		return
	}
	containers, err := s.Sys.Stores.Data.ListContainers(ctx)
	if err != nil {
		http.Error(w, "failed to load containers", http.StatusInternalServerError)
		return
	}

	// Merge live status for agents
	liveMap := s.Sys.Tracker.GetAgentMap()
	for _, a := range agents {
		if live, ok := liveMap[a.AgentID]; ok {
			a.Status = live.Status
			a.Since = live.Since
			a.UptimeSeconds = live.UptimeSeconds
			a.LastSeen = live.LastSeen
		} else {
			a.Status = "Offline"
			a.Since = "—"
		}
	}

	// Apply filtering
	filteredAgents := FilterAgents(agents, filter)
	filteredContainers := FilterContainers(containers, filter)

	// Convert to generic structure
	var result []map[string]interface{}

	for _, a := range filteredAgents {
		result = append(result, map[string]interface{}{
			"type":      "host",
			"id":        a.EndpointID,
			"hostname":  a.Hostname,
			"status":    a.Status,
			"last_seen": a.LastSeen,
			"uptime":    a.UptimeSeconds,
			"agent_id":  a.AgentID,
			"host_id":   a.HostID,
			"labels":    a.Labels,
			"ip":        a.IP,
			"os":        a.OS,
			"arch":      a.Arch,
			"version":   a.Version,
		})
	}

	for _, c := range filteredContainers {
		result = append(result, map[string]interface{}{
			"type":         "container",
			"id":           c.EndpointID,
			"container_id": c.ContainerID,
			"name":         c.Name,
			"image":        c.ImageName,
			"image_id":     c.ImageID,
			"runtime":      c.Runtime,
			"status":       c.Status,
			"host_id":      c.HostID,
			"last_seen":    c.LastSeen,
			"labels":       c.Labels,
		})
	}

	// Optional: sort by last_seen descending
	sort.SliceStable(result, func(i, j int) bool {
		t1, ok1 := result[i]["last_seen"].(time.Time)
		t2, ok2 := result[j]["last_seen"].(time.Time)
		return ok1 && ok2 && t1.After(t2)
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// FilterAgents filters a list of agents based on the provided filter criteria.
// It checks for matching endpoint ID, hostname, status, host ID, tags, and last seen time.
// The filtered list is then sorted and limited based on the specified criteria.
func FilterAgents(list []*model.Agent, filter EndpointFilter) []*model.Agent {
	var out []*model.Agent
	now := time.Now()

	for _, a := range list {
		if filter.EndpointID != "" && a.EndpointID != filter.EndpointID {
			continue
		}
		if filter.Hostname != "" && a.Hostname != filter.Hostname {
			continue
		}
		if filter.Status != "" && strings.ToLower(a.Status) != strings.ToLower(filter.Status) {
			continue
		}
		if filter.HostID != "" && a.HostID != filter.HostID {
			continue
		}
		if filter.IP != "" && a.IP != filter.IP {
			continue
		}
		if filter.OS != "" && a.OS != filter.OS {
			continue
		}
		if filter.Arch != "" && a.Arch != filter.Arch {
			continue
		}
		if !utils.MatchAllTags(filter.Tags, a.Labels) {
			continue
		}
		age := now.Sub(a.LastSeen)
		if filter.LastSeenMin > 0 && age < filter.LastSeenMin {
			continue
		}
		if filter.LastSeenMax > 0 && age > filter.LastSeenMax {
			continue
		}
		out = append(out, a)
	}

	return SortAndLimitAgents(out, filter.Sort, filter.Order, filter.Limit)
}

// HandleEndpointsByTypeAPI handles the /api/v1/endpoints/{endpointType} endpoint.
// It returns a JSON list of active endpoints filtered by the specified type (hosts or containers).
// It supports querying or listing all endpoints of the specified type.
// The endpoint type is specified in the URL path.
func (s *HttpServer) HandleEndpointsByTypeAPI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpointType := strings.ToLower(vars["endpointType"])

	ctx := r.Context()
	filter := ParseEndpointFilters(r)

	switch endpointType {
	case "hosts":
		agents, err := s.Sys.Stores.Data.ListAgents(ctx)
		if err != nil {
			http.Error(w, "failed to load agents", http.StatusInternalServerError)
			return
		}
		if agents == nil {
			agents = []*model.Agent{}
		}

		// Apply live status overlay
		liveMap := s.Sys.Tracker.GetAgentMap()
		for _, a := range agents {
			if live, ok := liveMap[a.AgentID]; ok {
				a.Status = live.Status
				a.Since = live.Since
				a.UptimeSeconds = live.UptimeSeconds
				a.LastSeen = live.LastSeen
			} else {
				a.Status = "Offline"
				a.Since = "—"
			}
		}

		filtered := FilterAgents(agents, filter)
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].Hostname < filtered[j].Hostname
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(filtered)

	case "containers":
		containers, err := s.Sys.Stores.Data.ListContainers(ctx)
		if err != nil {
			utils.Debug("Failed to load containers: %v", err)
			http.Error(w, "failed to load containers", http.StatusInternalServerError)
			return
		}
		if containers == nil {
			containers = []*model.Container{}
		}

		filtered := FilterContainers(containers, filter)
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].Name < filtered[j].Name
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(filtered)

	default:
		http.Error(w, "invalid endpoint type", http.StatusBadRequest)
	}
}

// FilterContainers filters a list of containers based on the provided filter criteria.
// It checks for matching endpoint ID, hostname, status, host ID, tags, and last seen time.
// The filtered list is then sorted and limited based on the specified criteria.
func FilterContainers(list []*model.Container, filter EndpointFilter) []*model.Container {
	var out []*model.Container
	now := time.Now()

	for _, c := range list {
		if filter.EndpointID != "" && c.EndpointID != filter.EndpointID {
			continue
		}
		if filter.Hostname != "" && c.Labels["hostname"] != filter.Hostname {
			continue
		}
		if filter.Status != "" && strings.ToLower(c.Status) != strings.ToLower(filter.Status) {
			continue
		}
		if filter.HostID != "" && c.HostID != filter.HostID {
			continue
		}
		if !utils.MatchAllTags(filter.Tags, c.Labels) {
			continue
		}
		age := now.Sub(c.LastSeen)
		if filter.LastSeenMin > 0 && age < filter.LastSeenMin {
			continue
		}
		if filter.LastSeenMax > 0 && age > filter.LastSeenMax {
			continue
		}
		out = append(out, c)
	}

	return SortAndLimitContainers(out, filter.Sort, filter.Order, filter.Limit)
}

// SortAndLimitAgents sorts a list of agents based on the specified criteria and limits the number of results.
// It sorts by hostname or last seen time, and can reverse the order.
func SortAndLimitAgents(list []*model.Agent, sortBy, order string, limit int) []*model.Agent {
	switch sortBy {
	case "hostname":
		sort.SliceStable(list, func(i, j int) bool {
			return list[i].Hostname < list[j].Hostname
		})
	case "last_seen":
		sort.SliceStable(list, func(i, j int) bool {
			return list[i].LastSeen.Before(list[j].LastSeen)
		})
	}
	if order == "desc" {
		slices.Reverse(list)
	}
	if len(list) > limit {
		return list[:limit]
	}
	return list
}

// SortAndLimitContainers sorts a list of containers based on the specified criteria and limits the number of results.
// It sorts by name or last seen time, and can reverse the order.
func SortAndLimitContainers(list []*model.Container, sortBy, order string, limit int) []*model.Container {
	switch sortBy {
	case "name":
		sort.SliceStable(list, func(i, j int) bool {
			return list[i].Name < list[j].Name
		})
	case "last_seen":
		sort.SliceStable(list, func(i, j int) bool {
			return list[i].LastSeen.Before(list[j].LastSeen)
		})
	}
	if order == "desc" {
		slices.Reverse(list)
	}
	if len(list) > limit {
		return list[:limit]
	}
	return list
}
