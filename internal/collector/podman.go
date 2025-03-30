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
along with LeetScraper. If not, see https://www.gnu.org/licenses/.
*/

package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PodmanCollector struct {
	client *http.Client
	name   string
}

func NewPodmanCollector(socketPath string) *PodmanCollector {

	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}

	return &PodmanCollector{
		client: &http.Client{
			Transport: transport,
			Timeout:   3 * time.Second,
		},
		name: "container_podman",
	}
}

func (p *PodmanCollector) Name() string {
	return p.name
}

func (p *PodmanCollector) Collect() MetricResult {

	req, err := http.NewRequest("GET", "http://d/v4.0.0/libpod/containers/json?all=true", nil)
	if err != nil {
		return MetricResult{
			Name: p.name,
			Err:  fmt.Errorf("failed to create podman request: %w", err),
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return MetricResult{
			Name: p.name,
			Err:  fmt.Errorf("failed to perform podman request: %w", err),
		}
	}
	defer resp.Body.Close()

	var containers []struct {
		ID      string   `json:"Id"`
		Names   []string `json:"Names"`
		State   string   `json:"State"`
		Created string   `json:"Created"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return MetricResult{
			Name: p.name,
			Err:  fmt.Errorf("failed to decode podman response: %w", err),
		}
	}
	log.Printf("[podman] found %d containers", len(containers))

	metrics := make(map[string]float64)
	meta := make(map[string]string)
	now := time.Now()

	for _, c := range containers {
		if len(c.Names) == 0 {
			continue
		}
		name := strings.TrimPrefix(c.Names[0], "/")

		// Store container state as metadata
		meta[fmt.Sprintf("container_podman_state_%s", name)] = c.State

		// Calculate uptime only if running
		if c.State == "running" && c.Created != "" {
			unixSec, err := strconv.ParseInt(c.Created, 10, 64)
			if err == nil && unixSec > 0 {
				uptime := now.Sub(time.Unix(unixSec, 0)).Seconds()
				metrics[fmt.Sprintf("container_podman_uptime_seconds_%s", name)] = uptime
			}
		}

		// Fetch container stats
		statsURL := fmt.Sprintf("http://d/v4.0.0/libpod/containers/%s/stats?stream=false", c.ID)
		statsReq, err := http.NewRequest("GET", statsURL, nil)
		if err != nil {
			continue
		}
		statsResp, err := p.client.Do(statsReq)
		if err != nil || statsResp.StatusCode != http.StatusOK {
			if statsResp != nil {
				statsResp.Body.Close()
			}
			continue
		}

		var stats struct {
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
				Usage uint64 `json:"usage"`
				Limit uint64 `json:"limit"`
			} `json:"memory_stats"`

			Networks map[string]struct {
				RxBytes   uint64 `json:"rx_bytes"`
				TxBytes   uint64 `json:"tx_bytes"`
				RxPackets uint64 `json:"rx_packets"`
				TxPackets uint64 `json:"tx_packets"`
			} `json:"networks"`

			BlkioStats struct {
				IOServiceBytesRecursive []struct {
					Major uint64 `json:"major"`
					Minor uint64 `json:"minor"`
					Op    string `json:"op"`
					Value uint64 `json:"value"`
				} `json:"io_service_bytes_recursive"`
			} `json:"blkio_stats"`
		}

		if err := json.NewDecoder(statsResp.Body).Decode(&stats); err != nil {
			statsResp.Body.Close()
			continue
		}
		statsResp.Body.Close()

		// CPU (raw usage — not delta percent)
		cpuPercent := 0.0
		if stats.CPUStats.OnlineCPUs > 0 {
			// Divide nanoseconds by 1e9 to get seconds, times 100 for percent
			cpuPercent = float64(stats.CPUStats.CPUUsage.TotalUsage) / 1e9 / float64(stats.CPUStats.OnlineCPUs) * 100
		}

		// Memory
		memUsage := float64(stats.MemoryStats.Usage)
		memLimit := float64(stats.MemoryStats.Limit)
		memPercent := 0.0
		if memLimit > 0 {
			memPercent = (memUsage / memLimit) * 100
		}

		log.Printf("[podman] container %s: cpu=%.2f%% mem=%.2f%%", name, cpuPercent, memPercent)

		metrics[fmt.Sprintf("container_podman_cpu_percent_%s", name)] = cpuPercent
		metrics[fmt.Sprintf("container_podman_mem_percent_%s", name)] = memPercent
		metrics[fmt.Sprintf("container_podman_mem_usage_bytes_%s", name)] = memUsage

		for iface, net := range stats.Networks {
			metrics[fmt.Sprintf("container_podman_net_rx_bytes_%s_%s", name, iface)] = float64(net.RxBytes)
			metrics[fmt.Sprintf("container_podman_net_tx_bytes_%s_%s", name, iface)] = float64(net.TxBytes)
		}

		for _, blk := range stats.BlkioStats.IOServiceBytesRecursive {
			switch blk.Op {
			case "Read":
				metrics[fmt.Sprintf("container_podman_blkio_read_bytes_%s", name)] += float64(blk.Value)
			case "Write":
				metrics[fmt.Sprintf("container_podman_blkio_write_bytes_%s", name)] += float64(blk.Value)
			}
		}
	}

	return MetricResult{
		Name: p.name,
		Data: metrics,
		Meta: meta, // contains container state
		Err:  nil,
	}
}
