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

// gosight/agent/internal/collector/docker.go
// docker.go - collects metrics from Docker containers

package container

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
	"github.com/aaronlmathis/gosight/shared/model"
)

type DockerCollector struct {
	client *client.Client
}

func NewDockerCollector() *DockerCollector {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil
	}
	return &DockerCollector{client: cli}
}

func (c *DockerCollector) Name() string {
	return "docker"
}

func (c *DockerCollector) Collect(ctx context.Context) ([]model.Metric, error) {
	if c.client == nil {
		return nil, nil
	}

	containers, err := c.client.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var metrics []model.Metric

	for _, ctr := range containers {
		statsResp, err := c.client.ContainerStats(ctx, ctr.ID, false)
		if err != nil {
			continue
		}
		var stats types.StatsJSON
		err = json.NewDecoder(statsResp.Body).Decode(&stats)
		statsResp.Body.Close()
		if err != nil {
			continue
		}

		dims := map[string]string{
			"container_id": ctr.ID[:12],
			"name":         strings.TrimPrefix(ctr.Names[0], "/"),
			"image":        ctr.Image,
			"status":       ctr.State,
			"runtime":      "docker",
			"mount_count":  strconv.Itoa(len(ctr.Mounts)),
		}
		for k, v := range ctr.Labels {
			dims["label."+k] = v
		}
		if parts := strings.Split(ctr.Image, ":"); len(parts) == 2 {
			dims["container_version"] = parts[1]
		}

		inspected, err := c.client.ContainerInspect(ctx, ctr.ID)
		if err == nil && inspected.State != nil && inspected.State.Health != nil {
			dims["health_status"] = inspected.State.Health.Status
		}

		uptime := 0.0
		if strings.ToLower(ctr.State) == "running" && ctr.Created > 0 {
			startTime := time.Unix(ctr.Created, 0)
			uptime = now.Sub(startTime).Seconds()
			if uptime > 1e6 || uptime < 0 {
				uptime = 0
			}
		}
		running := 0.0
		if strings.ToLower(ctr.State) == "running" {
			running = 1.0
		}
		metrics = append(metrics,
			agentutils.Metric("Container", "Docker", "uptime_seconds", uptime, "gauge", "seconds", dims, now),
			agentutils.Metric("Container", "Docker", "running", running, "gauge", "bool", dims, now),
		)

		metrics = append(metrics, ExtractAllDockerMetrics(stats, dims, now)...) // full stat extraction

		// Calculate CPU percent and network rates
		cpuPercent := calculateCPUPercent(ctr.ID, stats.CPUStats.CPUUsage.TotalUsage, stats.CPUStats.SystemUsage, int(stats.CPUStats.OnlineCPUs))
		rxRate, txRate := calculateNetRate(ctr.ID, now, sumNetRxRawDocker(stats), sumNetTxRawDocker(stats))

		metrics = append(metrics,
			agentutils.Metric("Container", "Docker", "cpu_percent", cpuPercent, "gauge", "percent", dims, now),
			agentutils.Metric("Container", "Docker", "net_rx_rate_bytes", rxRate, "gauge", "bytes/s", dims, now),
			agentutils.Metric("Container", "Docker", "net_tx_rate_bytes", txRate, "gauge", "bytes/s", dims, now),
		)
	}

	return metrics, nil
}

func ExtractAllDockerMetrics(stats types.StatsJSON, dims map[string]string, ts time.Time) []model.Metric {
	var metrics []model.Metric

	cpu := stats.CPUStats
	metrics = append(metrics,
		agentutils.Metric("Container", "Docker", "cpu_total_usage", float64(cpu.CPUUsage.TotalUsage), "counter", "nanoseconds", dims, ts),
		agentutils.Metric("Container", "Docker", "cpu_system_usage", float64(cpu.SystemUsage), "counter", "nanoseconds", dims, ts),
		agentutils.Metric("Container", "Docker", "cpu_online_cpus", float64(cpu.OnlineCPUs), "gauge", "count", dims, ts),
		agentutils.Metric("Container", "Docker", "cpu_throttle_periods", float64(cpu.ThrottlingData.Periods), "counter", "count", dims, ts),
		agentutils.Metric("Container", "Docker", "cpu_throttled_periods", float64(cpu.ThrottlingData.ThrottledPeriods), "counter", "count", dims, ts),
		agentutils.Metric("Container", "Docker", "cpu_throttled_time", float64(cpu.ThrottlingData.ThrottledTime), "counter", "nanoseconds", dims, ts),
	)
	for i, v := range cpu.CPUUsage.PercpuUsage {
		dimsCPU := copyDims(dims)
		dimsCPU["cpu"] = fmt.Sprintf("cpu%d", i)
		metrics = append(metrics,
			agentutils.Metric("Container", "Docker", "cpu_percpu_usage", float64(v), "counter", "nanoseconds", dimsCPU, ts),
		)
	}

	mem := stats.MemoryStats
	metrics = append(metrics,
		agentutils.Metric("Container", "Docker", "mem_usage_bytes", float64(mem.Usage), "gauge", "bytes", dims, ts),
		agentutils.Metric("Container", "Docker", "mem_limit_bytes", float64(mem.Limit), "gauge", "bytes", dims, ts),
		agentutils.Metric("Container", "Docker", "mem_max_usage_bytes", float64(mem.MaxUsage), "gauge", "bytes", dims, ts),
	)
	for k, v := range mem.Stats {
		metrics = append(metrics,
			agentutils.Metric("Container", "Docker", "mem_"+normalizeKey(k), float64(v), "gauge", "bytes", dims, ts),
		)
	}

	var rxTotal, txTotal uint64
	for iface, net := range stats.Networks {
		dimsNet := copyDims(dims)
		dimsNet["interface"] = iface
		rxTotal += net.RxBytes
		txTotal += net.TxBytes
		metrics = append(metrics,
			agentutils.Metric("Container", "Docker", "net_rx_bytes", float64(net.RxBytes), "counter", "bytes", dimsNet, ts),
			agentutils.Metric("Container", "Docker", "net_tx_bytes", float64(net.TxBytes), "counter", "bytes", dimsNet, ts),
			agentutils.Metric("Container", "Docker", "net_rx_packets", float64(net.RxPackets), "counter", "count", dimsNet, ts),
			agentutils.Metric("Container", "Docker", "net_tx_packets", float64(net.TxPackets), "counter", "count", dimsNet, ts),
			agentutils.Metric("Container", "Docker", "net_rx_errors", float64(net.RxErrors), "counter", "count", dimsNet, ts),
			agentutils.Metric("Container", "Docker", "net_tx_errors", float64(net.TxErrors), "counter", "count", dimsNet, ts),
			agentutils.Metric("Container", "Docker", "net_rx_dropped", float64(net.RxDropped), "counter", "count", dimsNet, ts),
			agentutils.Metric("Container", "Docker", "net_tx_dropped", float64(net.TxDropped), "counter", "count", dimsNet, ts),
		)
	}
	metrics = append(metrics,
		agentutils.Metric("Container", "Docker", "net_rx_bytes_total", float64(rxTotal), "counter", "bytes", dims, ts),
		agentutils.Metric("Container", "Docker", "net_tx_bytes_total", float64(txTotal), "counter", "bytes", dims, ts),
	)

	for _, entry := range stats.BlkioStats.IoServiceBytesRecursive {
		dimsBlk := copyDims(dims)
		dimsBlk["op"] = strings.ToLower(entry.Op)
		dimsBlk["device"] = fmt.Sprintf("%d:%d", entry.Major, entry.Minor)
		metrics = append(metrics,
			agentutils.Metric("Container", "Docker", "blkio_service_bytes", float64(entry.Value), "counter", "bytes", dimsBlk, ts),
		)
	}

	metrics = append(metrics,
		agentutils.Metric("Container", "Docker", "pids_current", float64(stats.PidsStats.Current), "gauge", "count", dims, ts),
	)

	return metrics
}
