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

type PodmanStats struct {
	Read string `json:"read"`
	Name string `json:"name"`
	ID   string `json:"id"`

	CPUStats struct {
		CPUUsage struct {
			TotalUsage        uint64 `json:"total_usage"`
			UsageInKernelmode uint64 `json:"usage_in_kernelmode"`
			UsageInUsermode   uint64 `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     int    `json:"online_cpus"`
	} `json:"cpu_stats"`

	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`

	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`

	Networks map[string]struct {
		RxBytes uint64 `json:"rx_bytes"`
		TxBytes uint64 `json:"tx_bytes"`
	} `json:"networks"`
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

type PodmanInspect struct {
	State struct {
		StartedAt string `json:"StartedAt"`
	} `json:"State"`
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
			fmt.Printf("⚠️ Failed to inspect %s: %v\n", ctr.ID, err)
		} else {
			t, err := time.Parse(time.RFC3339Nano, inspect.State.StartedAt)
			if err != nil {
				fmt.Printf("⚠️ Invalid StartedAt for %s: %q\n", ctr.ID, inspect.State.StartedAt)
			} else {
				ctr.StartedAt = t
			}
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

		cpuPercent := calculateCPUPercent(stats)
		metrics = append(metrics, model.Metric{
			Namespace:  "Container/Podman",
			Name:       "container.cpu.percent",
			Timestamp:  now,
			Value:      cpuPercent,
			Unit:       "percent",
			Dimensions: dims,
		})

		metrics = append(metrics, model.Metric{
			Namespace: "Container/Podman",
			Name:      "container.mem.usage_bytes",
			Timestamp: now,
			Value:     float64(stats.MemoryStats.Usage),

			Unit:       "bytes",
			Dimensions: dims,
		})

		metrics = append(metrics, model.Metric{
			Namespace:  "Container/Podman",
			Name:       "container.mem.limit_bytes",
			Timestamp:  now,
			Value:      float64(stats.MemoryStats.Limit),
			Unit:       "bytes",
			Dimensions: dims,
		})

		var rxTotal, txTotal uint64
		for _, net := range stats.Networks {
			rxTotal += net.RxBytes
			txTotal += net.TxBytes
		}

		metrics = append(metrics, model.Metric{
			Namespace:  "Container/Podman",
			Name:       "container.net.rx_bytes",
			Timestamp:  now,
			Value:      float64(rxTotal),
			Unit:       "bytes",
			Dimensions: dims,
		})

		metrics = append(metrics, model.Metric{
			Namespace:  "Container/Podman",
			Name:       "container.net.tx_bytes",
			Timestamp:  now,
			Value:      float64(txTotal),
			Unit:       "bytes",
			Dimensions: dims,
		})
	}

	return metrics, nil
}

func inspectContainer(socketPath, containerID string) (*PodmanInspect, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("http://d/v4.5.0/containers/%s/json", containerID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var inspect PodmanInspect
	if err := json.NewDecoder(resp.Body).Decode(&inspect); err != nil {
		return nil, err
	}
	return &inspect, nil
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

func calculateCPUPercent(stats *PodmanStats) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(stats.CPUStats.SystemCPUUsage - stats.PreCPUStats.SystemCPUUsage)
	onlineCPUs := float64(stats.CPUStats.OnlineCPUs)

	if sysDelta > 0.0 && cpuDelta > 0.0 && onlineCPUs > 0.0 {
		return (cpuDelta / sysDelta) * onlineCPUs * 100.0
	}
	return 0.0
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
