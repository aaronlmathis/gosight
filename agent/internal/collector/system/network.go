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

	interfaces, err := net.IOCounters(true) // true = per interface
	if err != nil {
		utils.Error("❌ Failed to get network IO counters: %v", err)
		return nil, err
	}

	for _, iface := range interfaces {
		dimensions := map[string]string{
			"interface": iface.Name,
		}

		metrics = append(metrics,
			model.Metric{
				Namespace:  "System/Network",
				Name:       "network.bytes_sent_total",
				Timestamp:  now,
				Value:      float64(iface.BytesSent),
				Unit:       "bytes",
				Dimensions: dimensions,
			},
			model.Metric{
				Namespace:  "System/Network",
				Name:       "network.bytes_recv_total",
				Timestamp:  now,
				Value:      float64(iface.BytesRecv),
				Unit:       "bytes",
				Dimensions: dimensions,
			},
			model.Metric{
				Namespace:  "System/Network",
				Name:       "network.packets_sent_total",
				Timestamp:  now,
				Value:      float64(iface.PacketsSent),
				Unit:       "count",
				Dimensions: dimensions,
			},
			model.Metric{
				Namespace:  "System/Network",
				Name:       "network.packets_recv_total",
				Timestamp:  now,
				Value:      float64(iface.PacketsRecv),
				Unit:       "count",
				Dimensions: dimensions,
			},
			model.Metric{
				Namespace:  "System/Network",
				Name:       "network.err_in_total",
				Timestamp:  now,
				Value:      float64(iface.Errin),
				Unit:       "count",
				Dimensions: dimensions,
			},
			model.Metric{
				Namespace:  "System/Network",
				Name:       "network.err_out_total",
				Timestamp:  now,
				Value:      float64(iface.Errout),
				Unit:       "count",
				Dimensions: dimensions,
			},
		)
	}

	return metrics, nil
}
