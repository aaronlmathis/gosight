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

	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
	"github.com/aaronlmathis/gosight/shared/model"
)

type PodmanCollector struct {
	socketPath string
}

func NewPodmanCollector() *PodmanCollector {
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

var prevStats = map[string]struct {
	CPUUsage  uint64
	SystemCPU uint64
	NetRx     uint64
	NetTx     uint64
	Timestamp time.Time
}{}

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
		if err == nil {
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

		//utils.Debug("Is container_id set? %s", ctr.ID)
		dims := map[string]string{
			"container_id": ctr.ID[:12],
			"name":         strings.TrimPrefix(ctr.Names[0], "/"),
			"image":        ctr.Image,
			"status":       ctr.State,
		}
		for k, v := range ctr.Labels {
			dims["label."+k] = v
		}
		if ports := formatPorts(ctr.Ports); ports != "" {
			dims["ports"] = ports
		}
		cpuPercent := calculateCPUPercent(ctr.ID, stats)
		rxRate, txRate := calculateNetRate(ctr.ID, now, sumNetRxRaw(stats), sumNetTxRaw(stats))
		//utils.Debug("Container dimensions: %+v", dims)
		metrics = append(metrics,
			agentutils.Metric("Container", "Podman", "uptime_seconds", uptime, "gauge", "seconds", dims, now),
			agentutils.Metric("Container", "Podman", "running", running, "gauge", "bool", dims, now),
			agentutils.Metric("Container", "Podman", "cpu_percent", cpuPercent, "gauge", "percent", dims, now),
			agentutils.Metric("Container", "Podman", "mem_usage_bytes", float64(stats.MemoryStats.Usage), "gauge", "bytes", dims, now),
			agentutils.Metric("Container", "Podman", "mem_limit_bytes", float64(stats.MemoryStats.Limit), "gauge", "bytes", dims, now),
			agentutils.Metric("Container", "Podman", "net_rx_bytes", rxRate, "gauge", "bytes", dims, now),
			agentutils.Metric("Container", "Podman", "net_tx_bytes", txRate, "gauge", "bytes", dims, now),
		)

	}

	return metrics, nil
}

/*
	func sumNetRx(stats *PodmanStats) float64 {
		var rx uint64
		for _, net := range stats.Networks {
			rx += net.RxBytes
		}
		return float64(rx)
	}

	func sumNetTx(stats *PodmanStats) float64 {
		var tx uint64
		for _, net := range stats.Networks {
			tx += net.TxBytes
		}
		return float64(tx)
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
*/
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
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var containers []PodmanContainer
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, err
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
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stats PodmanStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}
	return &stats, nil
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

func calculateCPUPercent(containerID string, stats *PodmanStats) float64 {
	now := time.Now()
	prev, ok := prevStats[containerID]
	currentCPU := stats.CPUStats.CPUUsage.TotalUsage
	currentSystem := stats.CPUStats.SystemCPUUsage

	var percent float64
	if ok {
		cpuDelta := float64(currentCPU - prev.CPUUsage)
		sysDelta := float64(currentSystem - prev.SystemCPU)
		if sysDelta > 0 && cpuDelta > 0 {
			percent = (cpuDelta / sysDelta) * float64(stats.CPUStats.OnlineCPUs) * 100.0
		}
	}

	// Update cache
	prevStats[containerID] = struct {
		CPUUsage  uint64
		SystemCPU uint64
		NetRx     uint64
		NetTx     uint64
		Timestamp time.Time
	}{
		CPUUsage:  currentCPU,
		SystemCPU: currentSystem,
		NetRx:     sumNetRxRaw(stats),
		NetTx:     sumNetTxRaw(stats),
		Timestamp: now,
	}

	return percent
}

func calculateNetRate(containerID string, now time.Time, rx, tx uint64) (float64, float64) {
	prev, ok := prevStats[containerID]
	if !ok || prev.Timestamp.IsZero() {
		return 0, 0
	}
	seconds := now.Sub(prev.Timestamp).Seconds()
	if seconds <= 0 {
		return 0, 0
	}
	rxRate := float64(rx-prev.NetRx) / seconds
	txRate := float64(tx-prev.NetTx) / seconds
	return rxRate, txRate
}

func sumNetRxRaw(stats *PodmanStats) uint64 {
	var rx uint64
	for _, net := range stats.Networks {
		rx += net.RxBytes
	}
	return rx
}

func sumNetTxRaw(stats *PodmanStats) uint64 {
	var tx uint64
	for _, net := range stats.Networks {
		tx += net.TxBytes
	}
	return tx
}
