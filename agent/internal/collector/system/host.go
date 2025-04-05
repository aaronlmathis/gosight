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

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/shirou/gopsutil/v4/host"
)

type HostCollector struct{}

func NewHostCollector() *HostCollector {
	return &HostCollector{}
}

func (c *HostCollector) Name() string {
	return "host"
}

func (c *HostCollector) Collect(ctx context.Context) ([]model.Metric, error) {
	var metrics []model.Metric
	now := time.Now()

	// --- Collect Core Host Information ---
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		utils.Error("Error getting host info: %v", err)
		// Host info is pretty fundamental, return error if failed
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	// --- Collect Logged-in User Information ---
	// Note: Getting users might fail due to permissions or unavailable utmp databases
	users, err := host.UsersWithContext(ctx)
	userCount := 0 // Default to 0 if error occurs
	if err != nil {
		utils.Warn("Error getting host users (continuing anyway): %v", err)
		// Continue collection even if users can't be determined
	} else {
		userCount = len(users)
	}

	// --- Create Metrics ---

	// Uptime Metric
	metrics = append(metrics, model.Metric{
		Namespace: "System",
		Name:      "host.uptime",
		Timestamp: now,
		Value:     float64(info.Uptime), // Uptime in seconds
		Unit:      "seconds",
		// Dimensions could be added here, but maybe keep them on host.info
		Dimensions: map[string]string{"hostname": info.Hostname}, // Or add hostname Dimension: map[string]string{"hostname": info.Hostname}?
	})

	// Process Count Metric
	metrics = append(metrics, model.Metric{
		Namespace:  "System",
		Name:       "host.procs",
		Timestamp:  now,
		Value:      float64(info.Procs), // Number of processes
		Unit:       "count",
		Dimensions: map[string]string{"hostname": info.Hostname},
	})

	// Logged-in User Count Metric
	metrics = append(metrics, model.Metric{
		Namespace:  "System",
		Name:       "host.users_loggedin",
		Timestamp:  now,
		Value:      float64(userCount), // Number of logged-in users
		Unit:       "count",
		Dimensions: map[string]string{"hostname": info.Hostname},
	})

	// Informational Metric with Host Details as Dimensions
	// This sends relatively static info periodically as dimensions on a simple metric
	hostInfoDimensions := map[string]string{
		"hostname":              info.Hostname,
		"os":                    info.OS,             // e.g., "linux", "darwin", "windows"
		"platform":              info.Platform,       // e.g., "ubuntu", "arch", "centos", "darwin", "windows"
		"platform_family":       info.PlatformFamily, // e.g., "debian", "rhel", "arch", "suse", "gentoo", "darwin", "windows"
		"platform_version":      info.PlatformVersion,
		"kernel_version":        info.KernelVersion,
		"kernel_arch":           info.KernelArch,
		"virtualization_system": info.VirtualizationSystem, // e.g., "kvm", "docker", "vmware", ""
		"virtualization_role":   info.VirtualizationRole,   // e.g., "guest", "host"
		"host_id":               info.HostID,               // Often UUID, persistent across reboots
	}

	metrics = append(metrics, model.Metric{
		Namespace:  "System",
		Name:       "host.info", // Metric name indicating this carries info
		Timestamp:  now,
		Value:      1,      // Constant value, focus is on dimensions
		Unit:       "info", // Custom unit perhaps
		Dimensions: hostInfoDimensions,
	})

	// Optionally add Boot Time metric
	// metrics = append(metrics, model.Metric{
	// 	Namespace:  "System",
	// 	Name:       "host.boot_time",
	// 	Timestamp:  now,
	// 	Value:      float64(info.BootTime), // Boot time as Unix timestamp
	// 	Unit:       "unix_timestamp",
	// 	Dimensions: map[string]string{},
	// })

	utils.Debug("Collected host metrics: %v", metrics)

	return metrics, nil
}
