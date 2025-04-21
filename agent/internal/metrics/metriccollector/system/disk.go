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

// gosight/agent/internal/collector/system/disk.go
// Package system provides collectors for system hardware (CPU/RAM/DISK/ETC)
// disk.go collects metrics on disk usage and info.
// It uses the gopsutil library to gather CPU metrics.

package system

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/shirou/gopsutil/v4/disk"
)

type DiskCollector struct{}

func NewDiskCollector() *DiskCollector {
	return &DiskCollector{}
}

func (c *DiskCollector) Name() string {
	return "disk"
}

func (c *DiskCollector) Collect(ctx context.Context) ([]model.Metric, error) {
	var metrics []model.Metric
	now := time.Now()

	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	for _, p := range partitions {
		// Platform-neutral filtering
		if runtime.GOOS != "windows" {
			// Linux/macOS: skip virtual, pseudo, temp filesystems
			if strings.HasPrefix(p.Mountpoint, "/sys") ||
				strings.HasPrefix(p.Mountpoint, "/proc") ||
				strings.HasPrefix(p.Mountpoint, "/run") ||
				p.Fstype == "tmpfs" || p.Fstype == "devtmpfs" || p.Fstype == "overlay" {
				continue
			}
		} else {
			// Windows: skip reserved/empty/mapped/unmounted drives
			if p.Fstype == "" || p.Mountpoint == "" {
				continue
			}
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage == nil {
			continue
		}

		dims := map[string]string{
			"mountpoint": p.Mountpoint,                            // e.g. "/", "/data", or "C:\"
			"device":     strings.TrimPrefix(p.Device, "\\\\.\\"), /* Windows-style */
			"fstype":     p.Fstype,
		}

		metrics = append(metrics,
			agentutils.Metric("System", "Disk", "total", usage.Total, "gauge", "bytes", dims, now),
			agentutils.Metric("System", "Disk", "used", usage.Used, "gauge", "bytes", dims, now),
			agentutils.Metric("System", "Disk", "free", usage.Free, "gauge", "bytes", dims, now),
			agentutils.Metric("System", "Disk", "used_percent", usage.UsedPercent, "gauge", "percent", dims, now),
			agentutils.Metric("System", "Disk", "inodes_total", usage.InodesTotal, "gauge", "count", dims, now),
			agentutils.Metric("System", "Disk", "inodes_used", usage.InodesUsed, "gauge", "count", dims, now),
			agentutils.Metric("System", "Disk", "inodes_free", usage.InodesFree, "gauge", "count", dims, now),
			agentutils.Metric("System", "Disk", "inodes_used_percent", usage.InodesUsedPercent, "gauge", "percent", dims, now),
		)

	}

	if ioCounters, err := disk.IOCounters(); err == nil {
		for device, io := range ioCounters {
			dims := map[string]string{
				"device":        device,
				"serial_number": io.SerialNumber,
			}

			metrics = append(metrics,
				agentutils.Metric("System", "DiskIO", "read_count", io.ReadCount, "counter", "count", dims, now),
				agentutils.Metric("System", "DiskIO", "write_count", io.WriteCount, "counter", "count", dims, now),
				agentutils.Metric("System", "DiskIO", "read_bytes", io.ReadBytes, "counter", "bytes", dims, now),
				agentutils.Metric("System", "DiskIO", "write_bytes", io.WriteBytes, "counter", "bytes", dims, now),
				agentutils.Metric("System", "DiskIO", "read_time", io.ReadTime, "counter", "milliseconds", dims, now),
				agentutils.Metric("System", "DiskIO", "write_time", io.WriteTime, "counter", "milliseconds", dims, now),
				agentutils.Metric("System", "DiskIO", "io_time", io.IoTime, "counter", "milliseconds", dims, now),
				agentutils.Metric("System", "DiskIO", "merged_read_count", io.MergedReadCount, "counter", "count", dims, now),
				agentutils.Metric("System", "DiskIO", "merged_write_count", io.MergedWriteCount, "counter", "count", dims, now),
				agentutils.Metric("System", "DiskIO", "weighted_io", io.WeightedIO, "counter", "milliseconds", dims, now),
			)
		}
	}

	return metrics, nil
}
