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
// server/internal/collector/container/helpers.go

package container

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// Common struct used by both Podman and Docker
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

type PortMapping struct {
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort"`
	Type        string `json:"Type"`
}

func formatPorts(ports []PortMapping) string {
	if len(ports) == 0 {
		return ""
	}
	var out []string
	for _, p := range ports {
		if p.PublicPort > 0 {
			out = append(out, fmt.Sprintf("%d:%d/%s", p.PublicPort, p.PrivatePort, p.Type))
		} else {
			out = append(out, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
		}
	}
	return strings.Join(out, ",")
}

// Reusable fetcher for container lists
func fetchContainersFromSocket[T any](socketPath, endpoint string) ([]T, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", "http://unix"+endpoint, nil)
	if err != nil {
		return nil, err
	}
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

// Reusable fetcher for container stats
func fetchContainerStatsFromSocket[T any](socketPath, statsEndpoint string) (T, error) {
	var result T
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", "http://unix"+statsEndpoint, nil)
	if err != nil {
		return result, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}

// ---- CPU + NET tracking

var prevStats = map[string]struct {
	CPUUsage  uint64
	SystemCPU uint64
	NetRx     uint64
	NetTx     uint64
	Timestamp time.Time
}{}

func calculateCPUPercent(containerID string, stats *PodmanStats) float64 {
	now := time.Now()
	prev, ok := prevStats[containerID]
	currentCPU := stats.CPUStats.CPUUsage.TotalUsage
	currentSystem := stats.CPUStats.SystemCPUUsage

	var percent float64
	if ok {
		cpuDelta := float64(currentCPU - prev.CPUUsage)
		sysDelta := float64(currentSystem - prev.SystemCPU)
		if sysDelta > 0 && cpuDelta > 0 && stats.CPUStats.OnlineCPUs > 0 {
			percent = (cpuDelta / sysDelta) * float64(stats.CPUStats.OnlineCPUs) * 100.0
		}
	}

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
	var total uint64
	for _, net := range stats.Networks {
		total += net.RxBytes
	}
	return total
}

func sumNetTxRaw(stats *PodmanStats) uint64 {
	var total uint64
	for _, net := range stats.Networks {
		total += net.TxBytes
	}
	return total
}
