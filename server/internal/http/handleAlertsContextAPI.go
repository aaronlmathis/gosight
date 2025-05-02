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

// server/internal/http/handleAlertsContextAPI.go

package http

import (
	"net/http"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/sys"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
)

type AlertContextResponse struct {
	Alert  model.AlertInstance `json:"alert"`
	Logs   []model.LogEntry    `json:"logs"`
	Events []model.EventEntry  `json:"events"`
}

func AlertContextHandler(sys *sys.SystemContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		alertID := mux.Vars(r)["id"]
		windowStr := r.URL.Query().Get("window")
		if windowStr == "" {
			windowStr = "30m"
		}
		window, err := time.ParseDuration(windowStr)
		if err != nil {
			http.Error(w, "invalid window", http.StatusBadRequest)
			return
		}

		// Lookup alert instance
		alert, err := sys.Stores.Alerts.GetByID(r.Context(), alertID)
		if err != nil {
			http.Error(w, "alert not found", http.StatusNotFound)
			return
		}

		// Calculate time window
		center := alert.FiredAt
		start := center.Add(-window)
		end := center.Add(window)

		// Determine filters
		var endpointID, target string
		if alert.EndpointID != "" {
			endpointID = alert.EndpointID
		}
		if alert.Target != "" {
			target = alert.Target
		}

		// Fetch logs
		logs, _ := sys.Stores.Logs.GetRecentLogs(r.Context(), model.LogFilter{
			Start:      start,
			End:        end,
			EndpointID: endpointID,
			Target:     target,
			Limit:      1000,
		})

		// Fetch events
		events, _ := sys.Stores.Events.GetRecentEvents(r.Context(), model.EventFilter{
			Start:      start,
			End:        end,
			EndpointID: endpointID,
			Target:     target,
			Limit:      1000,
		})

		// Respond
		resp := AlertContextResponse{
			Alert:  alert,
			Logs:   logs,
			Events: events,
		}
		utils.SendJSON(w, resp)
	}
}