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
	"net/http"

	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type MetricMetaHandler struct {
	Index *store.MetricIndex
}

func NewMetricMetaHandler(index *store.MetricIndex) *MetricMetaHandler {
	return &MetricMetaHandler{Index: index}
}

func (h *MetricMetaHandler) GetNamespaces(w http.ResponseWriter, r *http.Request) {
	utils.JSON(w, http.StatusOK, h.Index.GetNamespaces())
}

func (h *MetricMetaHandler) GetSubNamespaces(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		http.Error(w, "missing ?namespace", http.StatusBadRequest)
		return
	}
	utils.JSON(w, http.StatusOK, h.Index.GetSubNamespaces(ns))
}

func (h *MetricMetaHandler) GetMetricNames(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	sub := r.URL.Query().Get("sub")
	if ns == "" || sub == "" {
		http.Error(w, "missing ?namespace and ?sub", http.StatusBadRequest)
		return
	}
	utils.JSON(w, http.StatusOK, h.Index.GetMetricNames(ns, sub))
}

func (h *MetricMetaHandler) GetDimensions(w http.ResponseWriter, r *http.Request) {
	utils.JSON(w, http.StatusOK, h.Index.GetDimensions())
}
