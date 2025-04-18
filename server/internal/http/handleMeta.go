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

// File: server/internal/http/handleMeta.go
// Description: This file contains the HTTP handlers for the GoSight server's metadata API.
// It includes handlers for fetching namespaces, sub-namespaces, metric names, and dimensions.

package httpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
)

func (s *HttpServer) GetNamespaces(w http.ResponseWriter, r *http.Request) {
	utils.JSON(w, http.StatusOK, s.MetricIndex.GetNamespaces())
}

func (s *HttpServer) GetSubNamespaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns := strings.ToLower(vars["namespace"])
	if ns == "" {
		http.Error(w, "missing namespace in URL path", http.StatusBadRequest)
		return
	}
	utils.JSON(w, http.StatusOK, s.MetricIndex.GetSubNamespaces(ns))
}

func (s *HttpServer) GetMetricNames(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns := strings.ToLower(vars["namespace"])
	sub := strings.ToLower(vars["sub"])
	if ns == "" || sub == "" {
		http.Error(w, "missing namespace or subnamespace in URL path", http.StatusBadRequest)
		return
	}
	utils.JSON(w, http.StatusOK, s.MetricIndex.GetMetricNames(ns, sub))
}

func (s *HttpServer) GetDimensions(w http.ResponseWriter, r *http.Request) {
	utils.JSON(w, http.StatusOK, s.MetricIndex.GetDimensions())
}

func (s *HttpServer) GetMetricData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns := strings.ToLower(vars["namespace"])
	sub := strings.ToLower(vars["sub"])
	metric := strings.ToLower(vars["metric"])

	valid := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !valid.MatchString(ns) || !valid.MatchString(sub) || !valid.MatchString(metric) {
		utils.JSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid namespace, subnamespace, or metric name format"})
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	var start, end time.Time
	var err error
	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			utils.JSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("invalid start time: %v", err)})
			return
		}
	}
	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			utils.JSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("invalid end time: %v", err)})
			return
		}
	}

	filters := parseQueryFilters(r)
	fullMetric := fmt.Sprintf("%s.%s.%s", ns, sub, metric)

	if start.IsZero() && end.IsZero() {
		start = time.Now().Add(-time.Hour)
		end = time.Now()
	}

	points, err := s.MetricStore.QueryRange(fullMetric, start, end, filters)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to query range data: %v", err)})
		return
	}
	utils.JSON(w, http.StatusOK, points)
}

func (s *HttpServer) GetLatestValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns := strings.ToLower(vars["namespace"])
	sub := strings.ToLower(vars["sub"])
	metric := strings.ToLower(vars["metric"])

	fullMetric := fmt.Sprintf("%s.%s.%s", ns, sub, metric)
	filters := parseQueryFilters(r)

	rows, err := s.MetricStore.QueryInstant(fullMetric, filters)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to query latest value: %v", err)})
		return
	}
	if len(rows) == 0 {
		utils.JSON(w, http.StatusOK, []model.Point{})
		return
	}

	utils.JSON(w, http.StatusOK, model.Point{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Value:     rows[0].Value,
	})
}

func (s *HttpServer) HandleAPIQuery(w http.ResponseWriter, r *http.Request) {

	utils.Debug("Known dimensions: %+v", s.MetricIndex.GetDimensions())

	query := r.URL.Query()

	metricNames := query["metric"]

	// Optional time range
	startStr := query.Get("start")
	endStr := query.Get("end")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			http.Error(w, "invalid 'start' format (RFC3339)", http.StatusBadRequest)
			return
		}
	}
	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			http.Error(w, "invalid 'end' format (RFC3339)", http.StatusBadRequest)
			return
		}
	}

	// Build filters
	filters := make(map[string]string)
	for key, vals := range query {
		if key == "metric" || key == "start" || key == "end" || len(vals) == 0 {
			continue
		}
		filters[key] = vals[0]
	}

	utils.Debug(" Query Mode: metric=%q, start=%v, end=%v, filters=%+v", metricNames, start, end, filters)
	if len(filters) == 0 && len(metricNames) == 0 {
		http.Error(w, "must specify at least one filter or a metric name", http.StatusBadRequest)
		return
	}

	var result any

	switch {
	case len(metricNames) > 0 && !start.IsZero() && !end.IsZero():
		result, err = s.MetricStore.QueryMultiRange(metricNames, start, end, filters)

	case len(metricNames) > 0:
		result, err = s.MetricStore.QueryMultiInstant(metricNames, filters)

	case len(metricNames) == 0:
		// Power mode — return matching metrics across all known names
		utils.Debug("📡 Metric omitted — searching all available metrics")

		names := s.MetricIndex.FilterMetricNames(filters)
		utils.Debug("🧪 Filtered metric names: %v", names)
		if len(names) == 0 {
			http.Error(w, "no metrics matched filters", http.StatusNotFound)
			return
		}

		if !start.IsZero() && !end.IsZero() {
			result, err = s.MetricStore.QueryMultiRange(names, start, end, filters)
		} else {
			result, err = s.MetricStore.QueryMultiInstant(names, filters)
		}
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("query failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// helper
func parseQueryFilters(r *http.Request) map[string]string {
	filters := make(map[string]string)
	for key, values := range r.URL.Query() {
		if key == "start" || key == "end" || key == "latest" || key == "step" {
			continue
		}
		if len(values) == 1 {
			filters[key] = values[0]
		} else if len(values) > 1 {
			filters[key] = fmt.Sprintf("~^(%s)$", strings.Join(values, "|"))
		}
	}
	return filters
}

// ExportQueryHandler handles flexible label-based queries without requiring a metric name.
// Supports optional time range via start= and end= query params.
func (s *HttpServer) HandleExportQuery(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Extract label filters
	labels := make([]string, 0)
	for k, vals := range q {
		if len(vals) > 0 && k != "start" && k != "end" {
			labels = append(labels, fmt.Sprintf(`%s="%s"`, k, vals[0]))
		}
	}

	// Build match[] expression
	sort.Strings(labels)
	matchExpr := fmt.Sprintf("{%s}", strings.Join(labels, ","))

	params := url.Values{}
	params.Add("match[]", matchExpr)

	// Optional time range
	if start := q.Get("start"); start != "" {
		params.Add("start", start)
	}
	if end := q.Get("end"); end != "" {
		params.Add("end", end)
	}
	if _, ok := q["start"]; !ok {
		params.Add("start", fmt.Sprintf("%d", time.Now().Add(-5*time.Minute).Unix()))
	}
	if _, ok := q["end"]; !ok {
		params.Add("end", fmt.Sprintf("%d", time.Now().Unix()))
	}

	// Format: json or prom line format
	if format := q.Get("format"); format != "" {
		params.Add("format", format)
	}

	// Build final URL
	exportURL := fmt.Sprintf("%s?%s", s.Config.Storage.URL+"/api/v1/export", params.Encode())

	resp, err := http.Get(exportURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Export query failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
