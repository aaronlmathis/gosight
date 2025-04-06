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

// Basic Handler for http server
// server/internal/http/containers.go
// Package httpserver provides HTTP handlers for the GoSight server.

package httpserver

import (
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type ContainerMetrics struct {
	Host   string            `json:"host"`
	Name   string            `json:"name"`
	Image  string            `json:"image"`
	Status string            `json:"status"`
	CPU    *float64          `json:"cpu,omitempty"`
	Mem    *float64          `json:"mem,omitempty"`
	RX     *float64          `json:"rx,omitempty"`
	TX     *float64          `json:"tx,omitempty"`
	Uptime *float64          `json:"uptime,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
	Ports  string            `json:"ports,omitempty"`
}

type ContainerHandler struct {
	Store store.MetricStore
}

func (h *ContainerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	queries := map[string]string{
		"cpu":    `container.cpu.percent`,
		"mem":    `container.mem.usage_bytes`,
		"rx":     `container.net.rx_bytes`,
		"tx":     `container.net.tx_bytes`,
		"uptime": `container.uptime.seconds`,
		"status": `container.running`,
	}

	results := make(map[string]*ContainerMetrics)

	for metricKey, metricName := range queries {
		rows, err := h.Store.QueryInstant(metricName, map[string]string{
			"namespace": "container",
		})
		if err != nil {
			http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		for _, row := range rows {
			id := row.Tags["container_id"]
			if id == "" {
				continue
			}
			if _, ok := results[id]; !ok {
				results[id] = &ContainerMetrics{
					Host:   row.Tags["hostname"],
					Name:   row.Tags["container_name"],
					Image:  row.Tags["image"],
					Status: "stopped",
					Labels: make(map[string]string),
					Ports:  row.Tags["ports"],
				}
				for k, v := range row.Tags {
					if strings.HasPrefix(k, "label.") {
						results[id].Labels[strings.TrimPrefix(k, "label.")] = v
					}
				}
			}

			val := row.Value
			switch metricKey {
			case "cpu":
				results[id].CPU = &val
			case "mem":
				results[id].Mem = &val
			case "rx":
				results[id].RX = &val
			case "tx":
				results[id].TX = &val
			case "uptime":
				if val > 0 && val < 1e6 {
					results[id].Uptime = &val
				}
			case "status":
				if val > 0 {
					results[id].Status = "running"
				}
			}
		}
	}

	// Convert results map to list
	containerList := make([]*ContainerMetrics, 0, len(results))
	for _, c := range results {
		containerList = append(containerList, c)
	}

	// Filters
	subFilter := r.URL.Query().Get("subnamespace")
	hostFilter := r.URL.Query().Get("host")
	imageFilter := r.URL.Query().Get("image")
	statusFilter := r.URL.Query().Get("status")

	filtered := make([]*ContainerMetrics, 0, len(containerList))
	for _, c := range containerList {
		if hostFilter != "" && c.Host != hostFilter {
			continue
		}
		if imageFilter != "" && !strings.Contains(c.Image, imageFilter) {
			continue
		}
		if statusFilter != "" && c.Status != statusFilter {
			continue
		}
		if subFilter != "" && c.Labels["subnamespace"] != subFilter {
			continue
		}
		filtered = append(filtered, c)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Host == filtered[j].Host {
			return filtered[i].Name < filtered[j].Name
		}
		return filtered[i].Host < filtered[j].Host
	})

	utils.JSON(w, http.StatusOK, filtered)
}

func RenderContainersPage(w http.ResponseWriter, r *http.Request, templateDir, env string) {
	tmplPath := filepath.Join(templateDir, "containers.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		utils.Error("Template parse error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title": "Containers - GoSight",
		"Env":   env,
	}

	_ = tmpl.Execute(w, data)
}
