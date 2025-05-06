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

// File: server/internal/http/handleAlertContextAPI.go
package httpserver

import (
	"net/http"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
)

type AlertContextResponse struct {
	Alert  model.AlertInstance `json:"alert"`
	Logs   []model.LogEntry    `json:"logs"`
	Events []model.EventEntry  `json:"events"`
}

func (s *HttpServer) HandleAlertContext(w http.ResponseWriter, r *http.Request) {
	// Extract the alert ID from the URL parameters
	id := mux.Vars(r)["id"]
	utils.Debug("Fetching context for alert ID: %s", id)
	// Retrieve the alert by ID
	alert, err := s.Sys.Stores.Alerts.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "alert not found", http.StatusNotFound)
		return
	}

	// Default window duration is 30 minutes if not provided
	windowStr := r.URL.Query().Get("window")
	if windowStr == "" {
		windowStr = "30m"
	}

	// Parse the window duration
	window, err := time.ParseDuration(windowStr)
	if err != nil {
		http.Error(w, "invalid time window", http.StatusBadRequest)
		return
	}

	// Set the start and end time for the alert context window based on when the alert was fired
	start := alert.FirstFired.Add(-window) // Start time is 30 minutes before the first fired time
	end := alert.FirstFired.Add(window)    // End time is 30 minutes after the first fired time

	// Create a log filter based on the alert's context
	filter := model.LogFilter{
		Start:      start,
		End:        end,
		EndpointID: alert.EndpointID,
		Target:     alert.Target,
		Limit:      1000,  // Limit to 1000 logs
		Order:      "asc", // Order logs in ascending order by timestamp
	}

	// Fetch logs based on the filter
	logs, err := s.Sys.Stores.Logs.GetLogs(filter)
	if err != nil {
		http.Error(w, "failed to fetch logs", http.StatusInternalServerError)
		return
	}

	// Create an event filter based on the alert's context
	eventFilter := model.EventFilter{
		Start:      &start,
		End:        &end,
		EndpointID: alert.EndpointID,
		Target:     alert.Target,
		Limit:      1000,  // Limit to 1000 events
		SortOrder:  "asc", // Order events in ascending order by timestamp
	}

	// Fetch events based on the filter
	events, err := s.Sys.Stores.Events.GetRecentEvents(r.Context(), eventFilter)
	if err != nil {
		http.Error(w, "failed to fetch events", http.StatusInternalServerError)
		return
	}

	// Build the response structure
	resp := AlertContextResponse{
		Alert:  alert,  // Include the alert instance in the response
		Logs:   logs,   // Include the fetched logs in the response
		Events: events, // Include the fetched events in the response
	}

	// Send the response as JSON
	utils.JSON(w, 200, resp)
}
