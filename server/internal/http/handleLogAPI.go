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

// File: server/internal/http/handleLogAPI.go

// Description: This file contains the HTTP handlers for the GoSight server's log API.

package httpserver

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// HandleRecentLogs handles the HTTP request for recent logs.
// It retrieves the logs from the log store, applies any filters specified
// in the query parameters, and returns the logs as a JSON response.
// The limit for the number of logs returned can be specified in the query
// parameters, with a maximum of 1000 logs. If the limit is not specified,
// it defaults to 100 logs. The function also handles errors and returns
// appropriate HTTP status codes and messages.

func (s *HttpServer) HandleRecentLogs(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			if parsed > 1000 {
				parsed = 1000
			}
			limit = parsed
		}
	}

	filter := model.LogFilter{
		Limit: limit,
		Order: "desc",
	}

	logs, err := s.Sys.Stores.Logs.GetRecentLogs(r.Context(), filter)
	if err != nil {
		utils.Error("Failed to load logs: %v", err)
		http.Error(w, "failed to load logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(logs)
}

// HandleLogAPI handles the HTTP request for the log API.
// It retrieves the logs from the log store, applies any filters specified
// in the query parameters, and returns the logs as a JSON response.
// The function uses the LogQueryParams struct to parse the query parameters
// and filter the logs. It handles errors and returns appropriate HTTP status
// codes and messages. The logs are filtered based on the specified levels,
// unit, source, contains string, and start and end times. The function
// limits the number of logs returned to the specified limit in the query
// parameters, with a maximum of 1000 logs. If the limit is not specified,
// it defaults to 100 logs. The function also handles errors and returns
// appropriate HTTP status codes and messages.

func (s *HttpServer) HandleLogAPI(w http.ResponseWriter, r *http.Request) {
	filter := parseLogFilterFromQuery(r)

	logs, err := s.Sys.Stores.Logs.GetRecentLogs(r.Context(), filter)
	if err != nil {
		utils.Error("log query failed: %v", err)
		http.Error(w, "log query failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(logs)
}

// parseLogQueryParams parses the query parameters from the HTTP request
// and returns a LogQueryParams struct. It handles the limit, levels, unit,
// source, contains, start, and end parameters. The limit is capped at 1000.
// The levels are stored in a map for quick lookup. The start and end times
// are parsed as RFC3339 formatted strings and returned as pointers to time.Time.
// If a parameter is not provided or invalid, it is ignored.
// The function also trims whitespace and converts levels to lowercase.

func parseLogFilterFromQuery(r *http.Request) model.LogFilter {
	q := r.URL.Query()

	limit := 100
	if l, err := strconv.Atoi(q.Get("limit")); err == nil && l > 0 {
		if l > 1000 {
			limit = 1000
		} else {
			limit = l
		}
	}

	start := parseTime(q.Get("start"))
	end := parseTime(q.Get("end"))

	return model.LogFilter{
		Start:      start,
		End:        end,
		EndpointID: q.Get("endpointID"),
		Target:     q.Get("target"),
		Level:      strings.ToLower(strings.TrimSpace(q.Get("level"))),
		Category:   q.Get("unit"),
		Source:     q.Get("source"),
		Contains:   q.Get("keyword"),
		Limit:      limit,
		Order:      "desc",
	}
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02T15:04:05", s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// matchesSearch checks if a log entry matches the search keyword.
// It checks if the keyword is present in the message, source, category,
// or any of the meta fields, tags, or fields of the log entry.
// The function is case-insensitive and uses strings.Contains to check for matches.
// It returns true if the log entry matches the search keyword, false otherwise.
func matchesSearch(log model.LogEntry, keyword string) bool {
	kw := strings.ToLower(keyword)

	for k, v := range log.Meta.Extra {
		// already safe because map, but good habit:
		if strings.Contains(strings.ToLower(k), kw) || strings.Contains(strings.ToLower(v), kw) {
			return true
		}
	}

	return strings.Contains(strings.ToLower(log.Message), kw) ||
		strings.Contains(strings.ToLower(log.Source), kw) ||
		strings.Contains(strings.ToLower(log.Category), kw) ||
		strings.Contains(strings.ToLower(log.Level), kw) ||

		strings.Contains(strings.ToLower(log.Meta.Platform), kw) ||
		strings.Contains(strings.ToLower(log.Meta.AppName), kw) ||
		strings.Contains(strings.ToLower(log.Meta.AppVersion), kw) ||
		strings.Contains(strings.ToLower(log.Meta.ContainerID), kw) ||
		strings.Contains(strings.ToLower(log.Meta.ContainerName), kw) ||
		strings.Contains(strings.ToLower(log.Meta.Unit), kw) ||
		strings.Contains(strings.ToLower(log.Meta.Service), kw) ||
		strings.Contains(strings.ToLower(log.Meta.EventID), kw) ||
		strings.Contains(strings.ToLower(log.Meta.User), kw) ||
		strings.Contains(strings.ToLower(log.Meta.Executable), kw) ||
		strings.Contains(strings.ToLower(log.Meta.Path), kw)

}
