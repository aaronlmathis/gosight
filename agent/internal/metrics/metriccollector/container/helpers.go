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

	"github.com/docker/docker/api/types"
)

// fetchGenericJSON fetches JSON data from a Unix socket
// and decodes it into the provided target interface.
// It uses the provided socket path and endpoint to make the request.
// The function is generic and can return any type of data.
func fetchGenericJSON(socketPath, endpoint string, target interface{}) error {
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
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

// formatPorts formats a list of port mappings into a string
// representation. It handles both public and private ports.
// For example, it converts a list of PortMapping structs into
// a string like "8080:80/tcp,443/tcp".
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

// fetchContainerStatsFromSocket fetches container statistics from the Podman socket
// and returns the statistics for a specific container. It uses the
// provided socket path and endpoint to make the request. The function
// is generic and can return any type of container statistics.
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

// calculateCPUPercent calculates the CPU percentage for a container
// based on the total CPU usage and system CPU usage.
// It uses the previous CPU usage and system CPU usage to calculate
// the delta and then computes the percentage.
func calculateCPUPercent(containerID string, totalUsage, systemUsage uint64, onlineCPUs int) float64 {
	now := time.Now()
	prev, ok := prevStats[containerID]

	var percent float64
	if ok {
		cpuDelta := float64(totalUsage - prev.CPUUsage)
		sysDelta := float64(systemUsage - prev.SystemCPU)
		if sysDelta > 0 && cpuDelta > 0 && onlineCPUs > 0 {
			percent = (cpuDelta / sysDelta) * float64(onlineCPUs) * 100.0
		}
	}

	prevStats[containerID] = struct {
		CPUUsage  uint64
		SystemCPU uint64
		NetRx     uint64
		NetTx     uint64
		Timestamp time.Time
	}{
		CPUUsage:  totalUsage,
		SystemCPU: systemUsage,
		NetRx:     0,
		NetTx:     0,
		Timestamp: now,
	}

	return percent
}

// calculateNetRate calculates the network rate for a container
// based on the received and transmitted bytes.
// It uses the previous received and transmitted bytes to calculate
// the delta and then computes the rate in bytes per second.
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

	// update previous values
	prevStats[containerID] = struct {
		CPUUsage  uint64
		SystemCPU uint64
		NetRx     uint64
		NetTx     uint64
		Timestamp time.Time
	}{
		CPUUsage:  prev.CPUUsage,
		SystemCPU: prev.SystemCPU,
		NetRx:     rx,
		NetTx:     tx,
		Timestamp: now,
	}

	return rxRate, txRate
}

// sumNetRxRaw sums the received bytes from all network interfaces
// in the PodmanStats struct. It returns the total received bytes.
// It is used to calculate the network RX rate.
func sumNetRxRaw(stats *PodmanStats) uint64 {
	var total uint64
	for _, net := range stats.Networks {
		total += net.RxBytes
	}
	return total
}

// sumNetTxRaw sums the transmitted bytes from all network interfaces
// in the PodmanStats struct. It returns the total transmitted bytes.
// It is used to calculate the network TX rate.
func sumNetTxRaw(stats *PodmanStats) uint64 {
	var total uint64
	for _, net := range stats.Networks {
		total += net.TxBytes
	}
	return total
}

// sumNetRxRawDocker sums the received bytes from all network interfaces
// in the Docker Stats struct. It returns the total received bytes.
// It is used to calculate the network RX rate.
func copyDims(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// normalizeKey normalizes a string key by converting it to lowercase
// and replacing spaces with underscores. This is useful for
// standardizing keys in metrics and dimensions.
func normalizeKey(k string) string {
	return strings.ReplaceAll(strings.ToLower(k), " ", "_")
}

// sumNetRxRawDocker sums the received bytes from all network interfaces
// in the Docker Stats struct. It returns the total received bytes.
// It is used to calculate the network RX rate.
func sumNetRxRawDocker(stats types.StatsJSON) uint64 {
	var total uint64
	for _, iface := range stats.Networks {
		total += iface.RxBytes
	}
	return total
}

// sumNetTxRawDocker sums the transmitted bytes from all network interfaces
// in the Docker Stats struct. It returns the total transmitted bytes.
// It is used to calculate the network TX rate.
func sumNetTxRawDocker(stats types.StatsJSON) uint64 {
	var total uint64
	for _, iface := range stats.Networks {
		total += iface.TxBytes
	}
	return total
}
