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
// server/internal/collector/container/podman.go
// Package container provides a collector for Podman containers.
// It implements the Collector interface and collects metrics related to Podman containers.
package container

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
)

type PodmanCollector struct{}

func NewPodmanCollector() *PodmanCollector {
	return &PodmanCollector{}
}

func (c *PodmanCollector) Name() string {
	return "podman"
}

func (c *PodmanCollector) Collect(ctx context.Context) ([]model.Metric, error) {
	containers, err := fetchContainers()
	if err != nil {
		return nil, err
	}

	var metrics []model.Metric
	now := time.Now()

	for _, ctr := range containers {
		stats, err := fetchContainerStats(ctr.ID)
		if err != nil {
			continue
		}

		uptime := now.Sub(ctr.StartedAt).Seconds()
		running := 0.0
		if strings.ToLower(ctr.State) == "running" {
			running = 1.0
		}

		dims := map[string]string{
			"container_id": ctr.ID[:12],
			"name":         strings.TrimPrefix(ctr.Names[0], "/"),
			"image":        ctr.Image,
			"status":       ctr.State,
		}

		metrics = append(metrics, model.Metric{
			Namespace:  "Container/Podman",
			Name:       "container.uptime.seconds",
			Timestamp:  now,
			Value:      uptime,
			Unit:       "seconds",
			Dimensions: dims,
		})

		metrics = append(metrics, model.Metric{
			Namespace:  "Container/Podman",
			Name:       "container.running",
			Timestamp:  now,
			Value:      running,
			Unit:       "bool",
			Dimensions: dims,
		})

		metrics = append(metrics, model.Metric{
			Namespace:  "Container/Podman",
			Name:       "container.cpu.percent",
			Timestamp:  now,
			Value:      stats.CPU.Percent,
			Unit:       "percent",
			Dimensions: dims,
		})

		metrics = append(metrics, model.Metric{
			Namespace:  "Container/Podman",
			Name:       "container.mem.usage_bytes",
			Timestamp:  now,
			Value:      stats.Mem.Usage,
			Unit:       "bytes",
			Dimensions: dims,
		})

		metrics = append(metrics, model.Metric{
			Namespace:  "Container/Podman",
			Name:       "container.mem.limit_bytes",
			Timestamp:  now,
			Value:      stats.Mem.Limit,
			Unit:       "bytes",
			Dimensions: dims,
		})

		metrics = append(metrics, model.Metric{
			Namespace:  "Container/Podman",
			Name:       "container.net.rx_bytes",
			Timestamp:  now,
			Value:      stats.Net.RxBytes,
			Unit:       "bytes",
			Dimensions: dims,
		})

		metrics = append(metrics, model.Metric{
			Namespace:  "Container/Podman",
			Name:       "container.net.tx_bytes",
			Timestamp:  now,
			Value:      stats.Net.TxBytes,
			Unit:       "bytes",
			Dimensions: dims,
		})
	}

	return metrics, nil
}

func fetchContainers() ([]PodmanContainer, error) {
	sockPath := "/run/podman/podman.sock"
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sockPath)
			},
		},
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", "http://d/v4.0.0/containers/json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("podman API call failed: %w", err)
	}
	defer resp.Body.Close()

	var containers []PodmanContainer
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return containers, nil
}

func fetchContainerStats(containerID string) (*PodmanStats, error) {
	sockPath := "/run/podman/podman.sock"
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sockPath)
			},
		},
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("http://d/v4.0.0/containers/%s/stats?stream=false", containerID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create stats request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("podman stats API call failed: %w", err)
	}
	defer resp.Body.Close()

	var stats PodmanStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode stats response: %w", err)
	}

	return &stats, nil
}

// Minimal container struct from Podman API
type PodmanContainer struct {
	ID        string    `json:"Id"`
	Names     []string  `json:"Names"`
	Image     string    `json:"Image"`
	State     string    `json:"State"`
	StartedAt time.Time `json:"StartedAt"`
}

// Minimal stats struct from Podman API
type PodmanStats struct {
	CPU struct {
		Percent float64 `json:"cpu_percent"`
	} `json:"cpu_stats"`
	Mem struct {
		Usage float64 `json:"mem_usage"`
		Limit float64 `json:"mem_limit"`
	} `json:"mem_stats"`
	Net struct {
		RxBytes float64 `json:"rx_bytes"`
		TxBytes float64 `json:"tx_bytes"`
	} `json:"net_stats"`
}
