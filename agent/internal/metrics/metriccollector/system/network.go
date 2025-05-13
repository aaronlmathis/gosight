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

// gosight/agent/internal/collector/system/network.go
// GoSight - Network Collector
// Collects network interface I/O statistics via gopsutil

package system

import (
	"context"
	"time"

	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/shirou/gopsutil/v4/net"
)

type NetworkCollector struct{}

// NewNetworkCollector creates a new NetworkCollector instance.
// It initializes the collector and returns a pointer to it.
// This collector gathers network interface I/O statistics using the gopsutil library.
// It collects metrics such as bytes sent, bytes received, packets sent, packets received,
// and errors in/out for each network interface on the system.
func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{}
}

// Name returns the name of the collector.
// This is used to identify the collector in logs and metrics.
func (c *NetworkCollector) Name() string {
	return "network"
}

// Collect gathers network interface I/O statistics and returns them as a slice of model.Metric.
// It uses the gopsutil library to get network I/O counters for each interface.
// The metrics include bytes sent, bytes received, packets sent, packets received,
// and errors in/out for each network interface.
func (c *NetworkCollector) Collect(_ context.Context) ([]model.Metric, error) {
	now := time.Now()
	var metrics []model.Metric

	interfaces, err := net.IOCounters(true)
	if err != nil {
		utils.Error("❌ Failed to get network IO counters: %v", err)
		return nil, err
	}

	for _, iface := range interfaces {
		dims := map[string]string{"interface": iface.Name}

		metrics = append(metrics,
			agentutils.Metric("System", "Network", "bytes_sent", iface.BytesSent, "counter", "bytes", dims, now),
			agentutils.Metric("System", "Network", "bytes_recv", iface.BytesRecv, "counter", "bytes", dims, now),
			agentutils.Metric("System", "Network", "packets_sent", iface.PacketsSent, "counter", "count", dims, now),
			agentutils.Metric("System", "Network", "packets_recv", iface.PacketsRecv, "counter", "count", dims, now),
			agentutils.Metric("System", "Network", "err_in", iface.Errin, "counter", "count", dims, now),
			agentutils.Metric("System", "Network", "err_out", iface.Errout, "counter", "count", dims, now),
		)

	}

	return metrics, nil
}
