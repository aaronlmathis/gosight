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

// gosight/agent/internal/collector/system/host.go
// Package system provides collectors for system hardware (CPU/RAM/DISK/ETC)
// host.go collects metrics and information about the host system.
// It uses the gopsutil library to gather host metrics.

package system

import (
	"context"
	"fmt"
	"time"

	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/shirou/gopsutil/v4/host"
)

type HostCollector struct{}

// NewHostCollector creates a new HostCollector instance.
// It initializes the collector and returns a pointer to it.
// This collector gathers host system information such as uptime, number of processes,
// and number of logged-in users.
func NewHostCollector() *HostCollector {
	return &HostCollector{}
}

// Name returns the name of the collector.
// This is used to identify the collector in logs and metrics.
func (c *HostCollector) Name() string {
	return "host"
}

// Collect gathers host system information and returns it as a slice of model.Metric.
// It uses the gopsutil library to get host information such as uptime, number of processes,
// and number of logged-in users.
func (c *HostCollector) Collect(ctx context.Context) ([]model.Metric, error) {
	var metrics []model.Metric
	now := time.Now()

	info, err := host.InfoWithContext(ctx)
	if err != nil {
		utils.Error("Error getting host info: %v", err)
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	users, err := host.UsersWithContext(ctx)
	userCount := 0
	if err != nil {
		utils.Warn("Error getting host users (continuing anyway): %v", err)
	} else {
		userCount = len(users)
	}

	// Add simple numeric metrics
	metrics = append(metrics,
		agentutils.Metric("System", "Host", "uptime", info.Uptime, "gauge", "seconds", nil, now),
		agentutils.Metric("System", "Host", "procs", info.Procs, "gauge", "count", nil, now),
		agentutils.Metric("System", "Host", "users_loggedin", userCount, "gauge", "count", nil, now),
	)

	// Host info as dimension-only metric
	hostInfoDimensions := map[string]string{
		"hostname":              info.Hostname,
		"os":                    info.OS,
		"platform":              info.Platform,
		"platform_family":       info.PlatformFamily,
		"platform_version":      info.PlatformVersion,
		"kernel_version":        info.KernelVersion,
		"kernel_arch":           info.KernelArch,
		"virtualization_system": info.VirtualizationSystem,
		"virtualization_role":   info.VirtualizationRole,
		"host_id":               info.HostID,
	}

	metrics = append(metrics, agentutils.Metric("System", "Host", "info", 1, "gauge", "info", hostInfoDimensions, now))

	//utils.Debug("Collected host metrics: %v", metrics)
	return metrics, nil
}
