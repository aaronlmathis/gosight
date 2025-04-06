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

// gosight/agent/internal/http/metrics_meta.go
// Package httpserver provides HTTP handlers for the GoSight agent.
// It includes handlers for serving metric metadata, namespaces, and dimensions.

package httpserver

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
)

type MetricMetaHandler struct {
	Index *store.MetricIndex
	Store store.MetricStore
}

func NewMetricMetaHandler(index *store.MetricIndex, store store.MetricStore) *MetricMetaHandler {
	return &MetricMetaHandler{
		Index: index,
		Store: store,
	}
}

func (h *MetricMetaHandler) GetNamespaces(w http.ResponseWriter, r *http.Request) {
	utils.JSON(w, http.StatusOK, h.Index.GetNamespaces())
}

func (h *MetricMetaHandler) GetSubNamespaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns := strings.ToLower(vars["namespace"])

	if ns == "" {
		http.Error(w, "missing ?namespace", http.StatusBadRequest)
		return
	}
	utils.JSON(w, http.StatusOK, h.Index.GetSubNamespaces(ns))
}

func (h *MetricMetaHandler) GetMetricNames(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns := strings.ToLower(vars["namespace"])

	sub := strings.ToLower(vars["sub"])

	if ns == "" || sub == "" {
		http.Error(w, "missing namespace in the URL path", http.StatusBadRequest)
		return
	}
	utils.Debug("🔍 GetMetricNames: namespace=%s, sub=%s", ns, sub) // Add this log
	metricNames := h.Index.GetMetricNames(ns, sub)
	utils.Debug("🔍 GetMetricNames: Found metrics=%v", metricNames) // Add this log
	utils.JSON(w, http.StatusOK, metricNames)

}

func (h *MetricMetaHandler) GetDimensions(w http.ResponseWriter, r *http.Request) {
	utils.JSON(w, http.StatusOK, h.Index.GetDimensions())
}

func (h *MetricMetaHandler) GetMetricData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns := strings.ToLower(vars["namespace"])
	sub := strings.ToLower(vars["sub"])
	metric := strings.ToLower(vars["metric"])

	valid := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !valid.MatchString(ns) || !valid.MatchString(sub) || !valid.MatchString(metric) {
		utils.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid namespace, subnamespace, or metric name format",
		})
		return
	}

	fullMetricName := fmt.Sprintf("%s.%s.%s", ns, sub, metric)
	utils.Debug("📡 Querying metric data: %s", fullMetricName)

	// Optional time range
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			utils.JSON(w, http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("invalid start time: %v", err),
			})
			return
		}
	}
	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			utils.JSON(w, http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("invalid end time: %v", err),
			})
			return
		}
	}

	useAll := start.IsZero() && end.IsZero()

	if useAll {
		points, err := h.Store.QueryAll(fullMetricName)
		if err != nil {
			utils.JSON(w, http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("failed to query all data: %v", err),
			})
			return
		}
		utils.JSON(w, http.StatusOK, points)
		return
	}

	points, err := h.Store.QueryRange(fullMetricName, start, end)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("failed to query range data: %v", err),
		})
		return
	}
	utils.JSON(w, http.StatusOK, points)
}

func (h *MetricMetaHandler) GetLatestValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns := strings.ToLower(vars["namespace"])
	sub := strings.ToLower(vars["sub"])
	metric := strings.ToLower(vars["metric"])
	instance := r.URL.Query().Get("instance") // Get the 'instance' query parameter

	fullMetricName := fmt.Sprintf("%s.%s.%s", ns, sub, metric)
	utils.Debug("📡 Querying latest value for: %s", fullMetricName)

	rows, err := h.Store.QueryInstant(fullMetricName, instance)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("failed to query latest value: %v", err),
		})
		return
	}
	if len(rows) == 0 {
		utils.JSON(w, http.StatusOK, []model.Point{})
		return
	}

	// Use now as fallback, or get timestamp from VM result if available
	point := model.Point{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Value:     rows[0].Value,
	}
	utils.JSON(w, http.StatusOK, point)
}

func (h *MetricMetaHandler) HandleAPIQuery(w http.ResponseWriter, r *http.Request) {
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		utils.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "missing 'metric' parameter",
		})
		return
	}

	// Optional query modifiers
	latest := r.URL.Query().Get("latest") == "true"
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end time.Time
	var err error
	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			utils.JSON(w, http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("invalid start time: %v", err),
			})
			return
		}
	}
	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			utils.JSON(w, http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("invalid end time: %v", err),
			})
			return
		}
	}

	// Convert all remaining query parameters into Prometheus-style label filters
	matchers := make([]string, 0)
	for key, values := range r.URL.Query() {
		if key == "metric" || key == "start" || key == "end" || key == "latest" {
			continue
		}
		for _, val := range values {
			matchers = append(matchers, fmt.Sprintf(`%s="%s"`, key, val))
		}
	}
	query := metric
	if len(matchers) > 0 {
		query = fmt.Sprintf(`%s{%s}`, metric, strings.Join(matchers, ","))
	}

	if h.Store == nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": "metric store not available",
		})
		return
	}

	if latest {
		rows, err := h.Store.QueryInstant(query, "")
		if err != nil {
			utils.JSON(w, http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("query error: %v", err),
			})
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

	points, err := h.Store.QueryRange(query, start, end)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("query range error: %v", err),
		})
		return
	}

	utils.JSON(w, http.StatusOK, points)
}
