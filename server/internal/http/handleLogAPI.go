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

	logs, err := s.Sys.Stores.Logs.GetLogs(filter)
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
type LogResponse struct {
	Logs       []model.LogEntry `json:"logs"`
	NextCursor string           `json:"next_cursor,omitempty"`
	HasMore    bool             `json:"has_more"`
	Count      int              `json:"count"`
}

func (s *HttpServer) HandleLogAPI(w http.ResponseWriter, r *http.Request) {
	filter := parseLogFilterFromQuery(r)

	// Ensure a sane default
	if filter.Limit <= 0 || filter.Limit > 1000 {
		filter.Limit = 100
	}

	logs, err := s.Sys.Stores.Logs.GetLogs(filter)
	if err != nil {
		utils.Error("log query failed: %v", err)
		http.Error(w, "log query failed", http.StatusInternalServerError)
		return
	}

	// Pagination logic (cursor-based)
	var nextCursor string
	hasMore := false
	if len(logs) > filter.Limit {
		hasMore = true
		last := logs[filter.Limit-1]
		nextCursor = last.Timestamp.Format(time.RFC3339Nano)
		logs = logs[:filter.Limit] // trim to limit
	}

	resp := LogResponse{
		Logs:       logs,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Count:      len(logs),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
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

	parseTime := func(key string) time.Time {
		str := q.Get(key)
		if str == "" {
			return time.Time{}
		}
		t, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return time.Time{}
		}
		return t
	}

	parseInt := func(key string, def int) int {
		val := q.Get(key)
		if val == "" {
			return def
		}
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		return def
	}

	// Map extract helper
	extractPrefixed := func(prefix string) map[string]string {
		out := make(map[string]string)
		for key, values := range q {
			if strings.HasPrefix(key, prefix) && len(values) > 0 {
				k := strings.TrimPrefix(key, prefix)
				out[k] = values[0]
			}
		}
		return out
	}

	filter := model.LogFilter{
		Start:         parseTime("start"),
		End:           parseTime("end"),
		Cursor:        parseTime("cursor"),
		Limit:         parseInt("limit", 100),
		Order:         q.Get("order"),
		EndpointID:    q.Get("endpoint_id"),
		Target:        q.Get("target"),
		Level:         q.Get("level"),
		Category:      q.Get("category"),
		Source:        q.Get("source"),
		Contains:      q.Get("contains"),
		Unit:          q.Get("unit"),
		AppName:       q.Get("app_name"),
		Service:       q.Get("service"),
		EventID:       q.Get("event_id"),
		User:          q.Get("user"),
		ContainerID:   q.Get("container_id"),
		ContainerName: q.Get("container_name"),
		Platform:      q.Get("platform"),

		Tags:   extractPrefixed("tag_"),
		Fields: extractPrefixed("field_"),
		Meta:   extractPrefixed("meta_"),
	}

	return filter
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

	// Base fields
	if strings.Contains(strings.ToLower(log.Message), kw) ||
		strings.Contains(strings.ToLower(log.Source), kw) ||
		strings.Contains(strings.ToLower(log.Category), kw) ||
		strings.Contains(strings.ToLower(log.Level), kw) {
		return true
	}

	// Log.Meta fields
	if log.Meta != nil {
		metaFields := []string{
			log.Meta.Platform,
			log.Meta.AppName,
			log.Meta.AppVersion,
			log.Meta.ContainerID,
			log.Meta.ContainerName,
			log.Meta.Unit,
			log.Meta.Service,
			log.Meta.EventID,
			log.Meta.User,
			log.Meta.Executable,
			log.Meta.Path,
		}
		for _, f := range metaFields {
			if strings.Contains(strings.ToLower(f), kw) {
				return true
			}
		}

		for k, v := range log.Meta.Extra {
			if strings.Contains(strings.ToLower(k), kw) || strings.Contains(strings.ToLower(v), kw) {
				return true
			}
		}
	}

	// Tags and Fields
	for k, v := range log.Tags {
		if strings.Contains(strings.ToLower(k), kw) || strings.Contains(strings.ToLower(v), kw) {
			return true
		}
	}
	for k, v := range log.Fields {
		if strings.Contains(strings.ToLower(k), kw) || strings.Contains(strings.ToLower(v), kw) {
			return true
		}
	}

	return false
}
