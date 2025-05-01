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

// gosight/agent/internal/collector/registry.go
// registry.go - loads and initializes all enabled collectors at runtime.

package container

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
	"github.com/aaronlmathis/gosight/shared/model"
)

type PodmanCollector struct {
	SocketPath string
}

func NewPodmanCollector() *PodmanCollector {
	return &PodmanCollector{SocketPath: "/run/podman/podman.sock"}
}
func NewPodmanCollectorWithSocket(path string) *PodmanCollector {
	return &PodmanCollector{SocketPath: path}
}

func (c *PodmanCollector) Name() string {
	return "podman"
}

func (c *PodmanCollector) Collect(ctx context.Context) ([]model.Metric, error) {
	containers, err := fetchContainers[PodmanContainer](c.SocketPath, "/v4.0.0/containers/json?all=true")
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var metrics []model.Metric

	for _, ctr := range containers {
		stats, err := fetchStats(c.SocketPath, ctr.ID)
		if err != nil {
			continue
		}
		inspect, err := fetchInspect(c.SocketPath, ctr.ID)
		if err == nil && inspect.State.StartedAt != "" {
			t, err := time.Parse(time.RFC3339Nano, inspect.State.StartedAt)
			if err == nil {
				ctr.StartedAt = t
			}
		}

		uptime := 0.0
		if strings.ToLower(ctr.State) == "running" && !ctr.StartedAt.IsZero() {
			uptime = now.Sub(ctr.StartedAt).Seconds()
			if uptime > 1e6 || uptime < 0 {
				uptime = 0
			}
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
			"runtime":      "podman",
			"mount_count":  strconv.Itoa(len(ctr.Mounts)),
		}
		if parts := strings.Split(ctr.Image, ":"); len(parts) == 2 {
			dims["container_version"] = parts[1]
		}
		for k, v := range ctr.Labels {
			dims["label."+k] = v
		}
		if ports := formatPorts(ctr.Ports); ports != "" {
			dims["ports"] = ports
		}

		metrics = append(metrics,
			agentutils.Metric("Container", "Podman", "uptime_seconds", uptime, "gauge", "seconds", dims, now),
			agentutils.Metric("Container", "Podman", "running", running, "gauge", "bool", dims, now),
		)

		metrics = append(metrics, extractAllPodmanMetrics(stats, dims, now)...) // full stat extraction

		// Calculate CPU percent and network rates
		cpuPercent := calculateCPUPercent(ctr.ID, stats.CPUStats.CPUUsage.TotalUsage, stats.CPUStats.SystemCPUUsage, stats.CPUStats.OnlineCPUs)
		rxRate, txRate := calculateNetRate(ctr.ID, now, sumNetRxRaw(stats), sumNetTxRaw(stats))

		now := time.Now()
		metrics = append(metrics,
			agentutils.Metric("Container", "Podman", "cpu_percent", cpuPercent, "gauge", "percent", dims, now),
			agentutils.Metric("Container", "Podman", "net_rx_rate_bytes", rxRate, "gauge", "bytes/s", dims, now),
			agentutils.Metric("Container", "Podman", "net_tx_rate_bytes", txRate, "gauge", "bytes/s", dims, now),
		)
	}

	return metrics, nil
}

func extractAllPodmanMetrics(stats *PodmanStats, dims map[string]string, ts time.Time) []model.Metric {
	var metrics []model.Metric

	metrics = append(metrics,
		agentutils.Metric("Container", "Podman", "cpu_total_usage", float64(stats.CPUStats.CPUUsage.TotalUsage), "counter", "nanoseconds", dims, ts),
		agentutils.Metric("Container", "Podman", "cpu_kernelmode", float64(stats.CPUStats.CPUUsage.UsageInKernelmode), "counter", "nanoseconds", dims, ts),
		agentutils.Metric("Container", "Podman", "cpu_usermode", float64(stats.CPUStats.CPUUsage.UsageInUsermode), "counter", "nanoseconds", dims, ts),
		agentutils.Metric("Container", "Podman", "cpu_online_cpus", float64(stats.CPUStats.OnlineCPUs), "gauge", "count", dims, ts),
		agentutils.Metric("Container", "Podman", "cpu_system_usage", float64(stats.CPUStats.SystemCPUUsage), "counter", "nanoseconds", dims, ts),
	)

	metrics = append(metrics,
		agentutils.Metric("Container", "Podman", "mem_usage_bytes", float64(stats.MemoryStats.Usage), "gauge", "bytes", dims, ts),
		agentutils.Metric("Container", "Podman", "mem_limit_bytes", float64(stats.MemoryStats.Limit), "gauge", "bytes", dims, ts),
		agentutils.Metric("Container", "Podman", "mem_max_usage_bytes", 0, "gauge", "bytes", dims, ts),
	)

	var rx, tx uint64
	for iface, net := range stats.Networks {
		dimsNet := copyDims(dims)
		dimsNet["interface"] = iface
		rx += net.RxBytes
		tx += net.TxBytes
		metrics = append(metrics,
			agentutils.Metric("Container", "Podman", "net_rx_bytes", float64(net.RxBytes), "counter", "bytes", dimsNet, ts),
			agentutils.Metric("Container", "Podman", "net_tx_bytes", float64(net.TxBytes), "counter", "bytes", dimsNet, ts),
			agentutils.Metric("Container", "Podman", "net_rx_packets", 0, "counter", "count", dimsNet, ts),
			agentutils.Metric("Container", "Podman", "net_tx_packets", 0, "counter", "count", dimsNet, ts),
			agentutils.Metric("Container", "Podman", "net_rx_errors", 0, "counter", "count", dimsNet, ts),
			agentutils.Metric("Container", "Podman", "net_tx_errors", 0, "counter", "count", dimsNet, ts),
			agentutils.Metric("Container", "Podman", "net_rx_dropped", 0, "counter", "count", dimsNet, ts),
			agentutils.Metric("Container", "Podman", "net_tx_dropped", 0, "counter", "count", dimsNet, ts),
		)
	}
	metrics = append(metrics,
		agentutils.Metric("Container", "Podman", "net_rx_bytes_total", float64(rx), "counter", "bytes", dims, ts),
		agentutils.Metric("Container", "Podman", "net_tx_bytes_total", float64(tx), "counter", "bytes", dims, ts),
	)

	metrics = append(metrics,
		agentutils.Metric("Container", "Podman", "cpu_throttle_periods", 0, "counter", "count", dims, ts),
		agentutils.Metric("Container", "Podman", "cpu_throttled_periods", 0, "counter", "count", dims, ts),
		agentutils.Metric("Container", "Podman", "cpu_throttled_time", 0, "counter", "nanoseconds", dims, ts),
		agentutils.Metric("Container", "Podman", "pids_current", 0, "gauge", "count", dims, ts),
		agentutils.Metric("Container", "Podman", "blkio_service_bytes", 0, "counter", "bytes", dims, ts),
	)

	return metrics
}

func fetchContainers[T any](socketPath, endpoint string) ([]T, error) {
	client := &http.Client{Transport: unixTransport(socketPath), Timeout: 5 * time.Second}
	req, _ := http.NewRequest("GET", "http://unix"+endpoint, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out []T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func fetchStats(socketPath, containerID string) (*PodmanStats, error) {
	return fetchGeneric[PodmanStats](socketPath, fmt.Sprintf("/v4.0.0/containers/%s/stats?stream=false", containerID))
}

func fetchInspect(socketPath, containerID string) (*PodmanInspect, error) {
	return fetchGeneric[PodmanInspect](socketPath, fmt.Sprintf("/v4.5.0/containers/%s/json", containerID))
}

func fetchGeneric[T any](socketPath, endpoint string) (*T, error) {
	client := &http.Client{Transport: unixTransport(socketPath), Timeout: 5 * time.Second}
	req, _ := http.NewRequest("GET", "http://unix"+endpoint, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func unixTransport(socketPath string) *http.Transport {
	return &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}
}

type PodmanContainer struct {
	ID        string            `json:"Id"`
	Names     []string          `json:"Names"`
	Image     string            `json:"Image"`
	State     string            `json:"State"`
	Labels    map[string]string `json:"Labels"`
	Ports     []PortMapping     `json:"Ports"`
	Mounts    []any             `json:"Mounts"`
	StartedAt time.Time
}

type PodmanInspect struct {
	State struct {
		StartedAt string `json:"StartedAt"`
	} `json:"State"`
}

type PodmanStats struct {
	Read     string `json:"read"`
	Name     string `json:"name"`
	ID       string `json:"id"`
	CPUStats struct {
		CPUUsage struct {
			TotalUsage        uint64 `json:"total_usage"`
			UsageInKernelmode uint64 `json:"usage_in_kernelmode"`
			UsageInUsermode   uint64 `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     int    `json:"online_cpus"`
	} `json:"cpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage_bytes"`
		Limit uint64 `json:"limit_bytes"`
	} `json:"memory_stats"`
	Networks map[string]struct {
		RxBytes uint64 `json:"rx_bytes"`
		TxBytes uint64 `json:"tx_bytes"`
	} `json:"networks"`
}

type PortMapping struct {
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort"`
	Type        string `json:"Type"`
}

func dumpStatsRaw(socketPath, containerID string) {
	raw := make(map[string]interface{})
	err := fetchGenericJSON(socketPath, fmt.Sprintf("/v4.0.0/containers/%s/stats?stream=false", containerID), &raw)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Available Podman stats fields: %v\n", reflectKeys(raw))
}

func reflectKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
