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

// PodmanCollector collects metrics from Podman containers.
// It uses the Podman API to fetch container stats and metadata.
type PodmanCollector struct {
	SocketPath string
}

// PodmanContainer represents a Podman container.
// It contains metadata about the container, including its ID, name, image, state, labels, and ports.
// The ID is a unique identifier for the container.
// The Names field contains the names of the container.
// The Image field contains the name of the image used to create the container.
// The State field indicates the current state of the container (e.g., running, stopped).
// The Labels field contains key-value pairs of labels associated with the container.
// The Ports field contains a list of port mappings for the container.
// The Mounts field contains information about the mounts used by the container.
// The StartedAt field contains the time when the container was started.
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

// PodmanInspect represents the inspect data for a Podman container.
// It contains the state of the container, including the time when it was started.
// The State field contains information about the container's state.
// The StartedAt field contains the time when the container was started.
// The State field is a nested struct that contains the StartedAt field.
// The StartedAt field is a string in RFC3339 format.
type PodmanInspect struct {
	State struct {
		StartedAt string `json:"StartedAt"`
	} `json:"State"`
}

// PodmanStats represents the stats data for a Podman container.
// It contains information about CPU usage, memory usage, and network statistics.
// The CPUStats field contains information about CPU usage.
// The MemoryStats field contains information about memory usage.
// The Networks field contains information about network statistics.
// The CPUStats field is a nested struct that contains the CPUUsage field.
// The CPUUsage field contains information about CPU usage in kernel mode and user mode.

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

// PortMapping represents a port mapping for a Podman container.
// It contains the private port, public port, and type of mapping.
// The PrivatePort field contains the private port number.
// The PublicPort field contains the public port number.
// The Type field contains the type of mapping (e.g., tcp, udp).
type PortMapping struct {
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort"`
	Type        string `json:"Type"`
}

// NewPodmanCollector creates a new PodmanCollector with the default socket path.
// The default socket path is "/run/podman/podman.sock".
func NewPodmanCollector() *PodmanCollector {
	return &PodmanCollector{SocketPath: "/run/podman/podman.sock"}
}

// NewPodmanCollectorWithSocket creates a new PodmanCollector with a custom socket path.
// This is useful for testing or if the Podman socket is located in a different path.
// The socket path should be the full path to the Podman socket file.
func NewPodmanCollectorWithSocket(path string) *PodmanCollector {
	return &PodmanCollector{SocketPath: path}
}

// Name returns the name of the collector.
// This is used for logging and debugging purposes.
// It returns "podman" for the PodmanCollector.
func (c *PodmanCollector) Name() string {
	return "podman"
}

// Collect fetches container metrics from the Podman API.
// It returns a slice of model.Metric containing the collected metrics.
// If an error occurs during the collection process, it returns the error.
// The metrics include CPU usage, memory usage, network statistics, and container state.
func (c *PodmanCollector) Collect(_ context.Context) ([]model.Metric, error) {
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

// extractAllPodmanMetrics extracts all available metrics from the PodmanStats struct.
// It returns a slice of model.Metric containing the extracted metrics.
// The metrics include CPU usage, memory usage, network statistics, and other container stats.
// The metrics are tagged with the provided dimensions and timestamp.
// The dimensions are a map of key-value pairs that provide additional context for the metrics.
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

// fetchContainers fetches all containers from the Podman API.
// It returns a slice of PodmanContainer structs containing the container metadata.
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

// fetchStats fetches the stats for a specific container from the Podman API.
// It returns a PodmanStats struct containing the container stats.
func fetchStats(socketPath, containerID string) (*PodmanStats, error) {
	return fetchGeneric[PodmanStats](socketPath, fmt.Sprintf("/v4.0.0/containers/%s/stats?stream=false", containerID))
}

// fetchInspect fetches the inspect data for a specific container from the Podman API.
// It returns a PodmanInspect struct containing the container inspect data.
// The inspect data includes the container's state, labels, and other metadata.
func fetchInspect(socketPath, containerID string) (*PodmanInspect, error) {
	return fetchGeneric[PodmanInspect](socketPath, fmt.Sprintf("/v4.5.0/containers/%s/json", containerID))
}

// fetchGeneric fetches generic data from the Podman API.
// It takes a socket path and an endpoint as arguments.
// It returns a pointer to a generic type T and an error.
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

// unixTransport creates a new HTTP transport that uses a Unix socket.
// It takes a socket path as an argument and returns a pointer to http.Transport.
// This is used to communicate with the Podman API over a Unix socket.
func unixTransport(socketPath string) *http.Transport {
	return &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}
}
