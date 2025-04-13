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
	"fmt"
	"net/http"
	"sort"
	"strings"

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

func (s *HttpServer) HandleContainers(w http.ResponseWriter, r *http.Request) {
	queries := map[string]string{
		"cpu":    "cpu_percent",
		"mem":    "mem_usage_bytes",
		"rx":     "net_rx_bytes",
		"tx":     "net_tx_bytes",
		"uptime": "uptime_seconds",
		"status": "running",
	}

	results := make(map[string]*ContainerMetrics)
	subnamespaceSet := make(map[string]struct{}) // Track all container types (e.g. podman/docker)
	for metricKey, shortName := range queries {
		for _, sub := range []string{"podman", "docker"} { // TODO: Dynamically track this
			fullMetric := fmt.Sprintf("container.%s.%s", sub, shortName)
			rows, err := s.MetricStore.QueryInstant(fullMetric, map[string]string{
				"namespace": "container",
			})
			if err != nil {
				continue
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
				// Tag container type (subnamespace)
				if sub != "" {
					subnamespaceSet[sub] = struct{}{}
					results[id].Labels["subnamespace"] = sub
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
	}
	// Extract filters
	hostFilter := r.URL.Query().Get("host")
	imageFilter := r.URL.Query().Get("image")
	statusFilter := r.URL.Query().Get("status")
	subFilter := r.URL.Query().Get("subnamespace")

	filtered := make([]*ContainerMetrics, 0, len(results))
	for _, c := range results {
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
