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

func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{}
}

func (c *NetworkCollector) Name() string {
	return "network"
}

func (c *NetworkCollector) Collect(ctx context.Context) ([]model.Metric, error) {
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
