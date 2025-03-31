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

func sanitize(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
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
		ID           string            `json:"Id"`
		Names        []string          `json:"Names"`
		ImageName    string            `json:"Image"`
		State        string            `json:"State"`
		Created      string            `json:"Created"`
		Labels       map[string]string `json:"Labels"`
		RestartCount int               `json:"RestartCount"`
		Health       struct {
			Status string `json:"Status"`
		} `json:"HealthCheck"`
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
		meta[fmt.Sprintf("container_podman_image_%s", name)] = c.ImageName
		meta[fmt.Sprintf("container_podman_state_%s", name)] = c.State

		if c.ImageName != "" {
			meta[fmt.Sprintf("container_podman_image_%s", name)] = c.ImageName
		}

		meta[fmt.Sprintf("container_podman_restart_count_%s", name)] = strconv.Itoa(c.RestartCount)

		for k, v := range c.Labels {
			meta[fmt.Sprintf("container_podman_label_%s_%s", name, sanitize(k))] = v
		}

		// Calculate uptime only if running
		if c.State == "running" && c.Created != "" {
			unixSec, err := strconv.ParseInt(c.Created, 10, 64)
			if err == nil && unixSec > 0 {
				uptime := now.Sub(time.Unix(unixSec, 0)).Seconds()
				metrics[fmt.Sprintf("container_podman_uptime_seconds_%s", name)] = uptime
			}
		}

		// Fetch container stats
		// These are the main metrics that poulate Metrics in MetricResult

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

		/*******************************************************
			End Metrics
		***************************************************/

		// Begin Collecting the Meta data on containers.
		// This fills the Meta field of MetricResult
		// Fetch container details
		detailsReq, _ := http.NewRequest("GET", fmt.Sprintf("http://d/v4.0.0/libpod/containers/%s/json", c.ID), nil)
		detailsResp, err := p.client.Do(detailsReq)
		if err != nil {
			continue
		}
		defer detailsResp.Body.Close()

		var details struct {
			State struct {
				StartedAt string `json:"StartedAt"`
				Health    struct {
					Status string `json:"Status"`
				} `json:"Health"`
			} `json:"State"`
			Created string `json:"Created"`
			Image   string `json:"Image"`
			Config  struct {
				User         string                 `json:"User"`
				ExposedPorts map[string]interface{} `json:"ExposedPorts"`
			} `json:"Config"`
			HostConfig struct {
				RestartPolicy struct {
					Name              string `json:"Name"`
					MaximumRetryCount int    `json:"MaximumRetryCount"`
				} `json:"RestartPolicy"`
				NetworkMode string   `json:"NetworkMode"`
				Binds       []string `json:"Binds"`
			} `json:"HostConfig"`
		}

		log.Printf("[podman] requesting details for container %s", name)

		if err := json.NewDecoder(detailsResp.Body).Decode(&details); err != nil {
			log.Printf("[podman] decode error for %s: %v", name, err)
			continue
		}
		log.Printf("[podman] detail status code for %s: %d", name, detailsResp.StatusCode)

		// Health Status
		if hs := details.State.Health.Status; hs != "" {
			meta[fmt.Sprintf("container_podman_health_%s", name)] = hs
		}

		// Started At
		if details.State.StartedAt != "" {
			meta[fmt.Sprintf("container_podman_started_at_%s", name)] = details.State.StartedAt
		}

		// Created At
		if details.Created != "" {
			meta[fmt.Sprintf("container_podman_created_at_%s", name)] = details.Created
		}

		// Image Digest
		if details.Image != "" {
			meta[fmt.Sprintf("container_podman_image_digest_%s", name)] = details.Image
		}

		// Restart Policy
		if rp := details.HostConfig.RestartPolicy; rp.Name != "" {
			meta[fmt.Sprintf("container_podman_restart_policy_%s", name)] = rp.Name
		}
		if rp := details.HostConfig.RestartPolicy; rp.MaximumRetryCount > 0 {
			meta[fmt.Sprintf("container_podman_restart_max_%s", name)] = strconv.Itoa(rp.MaximumRetryCount)
		}

		// Network Mode
		if details.HostConfig.NetworkMode != "" {
			meta[fmt.Sprintf("container_podman_network_mode_%s", name)] = details.HostConfig.NetworkMode
		}

		// User
		if details.Config.User != "" {
			meta[fmt.Sprintf("container_podman_user_%s", name)] = details.Config.User
		}

		if c.State == "running" && details.Created != "" {
			createdTime, err := time.Parse(time.RFC3339Nano, details.Created)
			if err == nil {
				uptime := now.Sub(createdTime).Seconds()
				meta[fmt.Sprintf("container_podman_uptime_seconds_%s", name)] = fmt.Sprintf("%.2f", uptime)
			}
		}

		// Volumes
		if len(details.HostConfig.Binds) > 0 {
			meta[fmt.Sprintf("container_podman_volumes_%s", name)] = strings.Join(details.HostConfig.Binds, ",")
		}

		// Exposed Ports
		if len(details.Config.ExposedPorts) > 0 {
			ports := make([]string, 0, len(details.Config.ExposedPorts))
			for port := range details.Config.ExposedPorts {
				ports = append(ports, port)
			}
			meta[fmt.Sprintf("container_podman_ports_%s", name)] = strings.Join(ports, ",")
		}

	}

	return MetricResult{
		Name: p.name,
		Data: metrics,
		Meta: meta, // contains container state
		Err:  nil,
	}
}
