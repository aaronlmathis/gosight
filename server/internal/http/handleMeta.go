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
	"fmt"
	"net/http"
	"regexp"
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
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		utils.JSON(w, http.StatusBadRequest, map[string]string{"error": "missing 'metric' parameter"})
		return
	}

	latest := r.URL.Query().Get("latest") == "true"
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

	if latest {
		rows, err := s.MetricStore.QueryInstant(metric, filters)
		if err != nil {
			utils.JSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("query error: %v", err)})
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
		return
	}

	points, err := s.MetricStore.QueryRange(metric, start, end, filters)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("query range error: %v", err)})
		return
	}
	utils.JSON(w, http.StatusOK, points)
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
