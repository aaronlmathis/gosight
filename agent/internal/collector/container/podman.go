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

type PodmanCollector struct {
	socketPath string
}

func NewPodmanCollector() *PodmanCollector {
	// Default to rootful Podman socket
	return &PodmanCollector{socketPath: "/run/podman/podman.sock"}
}

func NewPodmanCollectorWithSocket(path string) *PodmanCollector {
	return &PodmanCollector{socketPath: path}
}

func (c *PodmanCollector) Name() string {
	return "podman"
}

type ContainerInspect struct {
	State struct {
		StartedAt string `json:"StartedAt"`
	} `json:"State"`
}

func inspectContainer(socketPath, id string) (*ContainerInspect, error) {
	url := fmt.Sprintf("http://d/v4.5.0/containers/%s/json", id)
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var inspect ContainerInspect
	if err := json.NewDecoder(resp.Body).Decode(&inspect); err != nil {
		return nil, err
	}
	return &inspect, nil
}
func (c *PodmanCollector) Collect(ctx context.Context) ([]model.Metric, error) {
	containers, err := fetchContainers(c.socketPath)
	if err != nil {
		return nil, err
	}

	var metrics []model.Metric
	now := time.Now()

	for _, ctr := range containers {
		stats, err := fetchContainerStats(c.socketPath, ctr.ID)
		if err != nil {
			continue
		}
		inspect, err := inspectContainer(c.socketPath, ctr.ID)
		if err != nil {
			fmt.Printf("⚠️  Failed to inspect container %s: %v\n", ctr.ID, err)
			continue
		}

		t, err := time.Parse(time.RFC3339Nano, inspect.State.StartedAt)
		if err != nil {
			fmt.Printf("⚠️  Invalid StartedAt for %s: %q\n", ctr.Image, inspect.State.StartedAt)
		} else {
			fmt.Printf("⏱️  %s started at %s (uptime %.1fs)\n", ctr.ID, t.Format(time.RFC3339), time.Since(t).Seconds())
			ctr.StartedAt = t
		}
		var uptime float64
		if !ctr.StartedAt.IsZero() {
			uptime = now.Sub(ctr.StartedAt).Seconds()
			if uptime > 1e6 || uptime < 0 {
				uptime = 0
			}
		} else {
			uptime = 0
		}
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

		// Enrich with labels and ports if available
		for k, v := range ctr.Labels {
			dims["label."+k] = v
		}

		if ports := formatPorts(ctr.Ports); ports != "" {
			dims["ports"] = ports
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

func fetchContainers(socketPath string) ([]PodmanContainer, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", "http://d/v4.0.0/containers/json?all=true", nil)
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

func fetchContainerStats(socketPath, containerID string) (*PodmanStats, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
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
	ID        string            `json:"Id"`
	Names     []string          `json:"Names"`
	Image     string            `json:"Image"`
	State     string            `json:"State"`
	StartedAt time.Time         `json:"StartedAt"`
	Labels    map[string]string `json:"Labels"`
	Ports     []PortMapping     `json:"Ports"`
}

type PortMapping struct {
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort"`
	Type        string `json:"Type"`
}

func formatPorts(ports []PortMapping) string {
	if len(ports) == 0 {
		return ""
	}
	var formatted []string
	for _, p := range ports {
		if p.PublicPort > 0 {
			formatted = append(formatted, fmt.Sprintf("%d:%d/%s", p.PublicPort, p.PrivatePort, p.Type))
		} else {
			formatted = append(formatted, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
		}
	}
	return strings.Join(formatted, ",")
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
