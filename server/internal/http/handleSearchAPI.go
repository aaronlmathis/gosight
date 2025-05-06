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

// gosight/agent/internal/http/handleSearchAPI.go

package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aaronlmathis/gosight/shared/model"
)

type SearchResult struct {
	Hosts      []map[string]string `json:"hosts"`
	Containers []map[string]string `json:"containers"`
	Rules      []map[string]string `json:"rules"`
	Tags       []map[string]string `json:"tags"`
	Logs       []map[string]string `json:"logs"`
}

func (s *HttpServer) HandleGlobalSearchAPI(w http.ResponseWriter, r *http.Request) {

	term := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("term")))
	if len(term) < 2 {
		http.Error(w, "search term too short", http.StatusBadRequest)
		return
	}

	result := SearchResult{
		Hosts:      []map[string]string{},
		Containers: []map[string]string{},
		Rules:      []map[string]string{},
		Tags:       []map[string]string{},
		Logs:       []map[string]string{},
	}

	// === 1. Fetch Endpoints (hosts + containers) ===
	endpoints := s.Sys.Tracker.ListEndpoints()
	hostsFull := false
	containersFull := false
	for _, ep := range endpoints {
		if strings.HasPrefix(ep.EndpointID, "host-") {
			// Host
			if strings.Contains(strings.ToLower(ep.Hostname), term) {
				result.Hosts = append(result.Hosts, map[string]string{
					"label":       ep.Hostname,
					"endpoint_id": ep.EndpointID,
				})
				if len(result.Hosts) >= 5 {
					hostsFull = true
				}
			}
		} else if strings.HasPrefix(ep.EndpointID, "ctr-") {
			// Container
			if strings.Contains(strings.ToLower(ep.ContainerName), term) {
				result.Containers = append(result.Containers, map[string]string{
					"label":       ep.ContainerName,
					"endpoint_id": ep.EndpointID,
				})
				if len(result.Containers) >= 5 {
					containersFull = true
				}
			}
		}

		if hostsFull && containersFull {
			break
		}
	}

	// === 2. Fetch Rules (alerts) ===
	alerts := s.Sys.Tele.Alerts.ListActive()

	for _, alert := range alerts {
		if strings.Contains(strings.ToLower(alert.RuleID), term) ||
			strings.Contains(strings.ToLower(alert.Message), term) {
			result.Rules = append(result.Rules, map[string]string{
				"label":   alert.RuleID,
				"rule_id": alert.RuleID,
			})
		}
		if len(result.Rules) >= 5 {
			break
		}
	}

	// === 3. Fetch Tags (keys + values) ===
	tagKeys, err := s.Sys.Stores.Data.ListKeys(r.Context())
	if err == nil {
		for _, key := range tagKeys {
			if strings.Contains(strings.ToLower(key), term) {
				result.Tags = append(result.Tags, map[string]string{
					"label": key,
				})
			}
			if len(result.Tags) >= 5 {
				break
			}
		}
	}
	filter := model.LogFilter{
		Limit: 100,
	}
	// === 4. Fetch Logs (recent messages) ===
	logs, err := s.Sys.Stores.Logs.GetLogs(filter) // Fetch last 100 logs
	if err == nil {
		for _, log := range logs {
			if strings.Contains(strings.ToLower(log.Message), term) {
				result.Logs = append(result.Logs, map[string]string{
					"label": log.Message,
				})
			}
			if len(result.Logs) >= 5 {
				break
			}
		}
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
