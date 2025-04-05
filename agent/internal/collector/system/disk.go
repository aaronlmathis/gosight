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
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
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

	// --- Collect Usage Metrics per Partition ---

	// Get all relevant partitions (physical devices, excludes pseudo filesystems)
	partitions, err := disk.Partitions(false) // Set to true if you want *everything* including tmpfs, devpts etc.
	if err != nil {
		utils.Error("Error getting disk partitions: %v", err)
		// Decide if you want to return here or try collecting IO counters anyway
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	for _, p := range partitions {
		// Get usage stats for this specific partition mount point
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			// Log error but continue with other partitions
			utils.Warn("Error getting disk usage for %s: %v", p.Mountpoint, err)
			continue // Skip this partition
		}

		if usage == nil {
			utils.Warn("Disk usage info is nil for %s", p.Mountpoint)
			continue // Skip this partition
		}

		// Add dimensions to identify the specific partition
		usageDimensions := map[string]string{
			"mountpoint": p.Mountpoint,
			"device":     strings.TrimPrefix(p.Device, "/dev/"), // e.g., /dev/sda1
			"fstype":     p.Fstype,
		}

		// Append various usage metrics for this partition
		metrics = append(metrics,
			model.Metric{
				Namespace:    "System",
				SubNamespace: "Disk",
				Name:         "disk.total",
				Timestamp:    now,
				Value:        float64(usage.Total), // Convert uint64 to float64
				Unit:         "bytes",
				Dimensions:   usageDimensions,
			},
			model.Metric{
				Namespace:    "System",
				SubNamespace: "Disk",
				Name:         "disk.used",
				Timestamp:    now,
				Value:        float64(usage.Used), // Convert uint64 to float64
				Unit:         "bytes",
				Dimensions:   usageDimensions,
			},
			model.Metric{
				Namespace:    "System",
				SubNamespace: "Disk",
				Name:         "disk.free",
				Timestamp:    now,
				Value:        float64(usage.Free), // Convert uint64 to float64
				Unit:         "bytes",
				Dimensions:   usageDimensions,
			},
			model.Metric{
				Namespace:    "System",
				SubNamespace: "Disk",
				Name:         "disk.used_percent",
				Timestamp:    now,
				Value:        usage.UsedPercent, // Already float64
				Unit:         "percent",
				Dimensions:   usageDimensions,
			},
			// Inode metrics (optional, but good to have)
			model.Metric{
				Namespace:    "System",
				SubNamespace: "Disk",
				Name:         "disk.inodes_total",
				Timestamp:    now,
				Value:        float64(usage.InodesTotal),
				Unit:         "count",
				Dimensions:   usageDimensions,
			},
			model.Metric{
				Namespace:    "System",
				SubNamespace: "Disk",
				Name:         "disk.inodes_used",
				Timestamp:    now,
				Value:        float64(usage.InodesUsed),
				Unit:         "count",
				Dimensions:   usageDimensions,
			},
			model.Metric{
				Namespace:    "System",
				SubNamespace: "Disk",
				Name:         "disk.inodes_free",
				Timestamp:    now,
				Value:        float64(usage.InodesFree),
				Unit:         "count",
				Dimensions:   usageDimensions,
			},
			model.Metric{
				Namespace:    "System",
				SubNamespace: "Disk",
				Name:         "disk.inodes_used_percent",
				Timestamp:    now,
				Value:        usage.InodesUsedPercent,
				Unit:         "percent",
				Dimensions:   usageDimensions,
			},
		)
	}

	// --- Collect I/O Counters per Device ---

	// Get IO statistics for block devices (e.g., sda, nvme0n1)
	// Passing no args gets all devices
	ioCounters, err := disk.IOCounters()
	if err != nil {
		// Log error but potentially return metrics collected so far
		utils.Error("Error getting disk IO counters: %v", err)
		// Decide if this error is critical enough to return, or just log
		// return metrics, fmt.Errorf("failed to get disk IO counters: %w", err) // Option 1: Return error
		// Option 2: Continue without IO counters (current behavior)
	} else {
		for deviceName, ioStat := range ioCounters {
			// Add dimensions to identify the specific device
			ioDimensions := map[string]string{
				"device":        deviceName, // e.g., sda, nvme0n1 (note: different from partition device name sometimes)
				"serial_number": ioStat.SerialNumber,
			}

			// Append various I/O metrics for this device
			metrics = append(metrics,
				model.Metric{
					Namespace:    "System",
					SubNamespace: "DiskIO",
					Name:         "diskio.read_count",
					Timestamp:    now,
					Value:        float64(ioStat.ReadCount),
					Unit:         "count",
					Dimensions:   ioDimensions,
				},
				model.Metric{
					Namespace:    "System",
					SubNamespace: "DiskIO",
					Name:         "diskio.write_count",
					Timestamp:    now,
					Value:        float64(ioStat.WriteCount),
					Unit:         "count",
					Dimensions:   ioDimensions,
				},
				model.Metric{
					Namespace:    "System",
					SubNamespace: "DiskIO",
					Name:         "diskio.read_bytes",
					Timestamp:    now,
					Value:        float64(ioStat.ReadBytes),
					Unit:         "bytes",
					Dimensions:   ioDimensions,
				},
				model.Metric{
					Namespace:    "System",
					SubNamespace: "DiskIO",
					Name:         "diskio.write_bytes",
					Timestamp:    now,
					Value:        float64(ioStat.WriteBytes),
					Unit:         "bytes",
					Dimensions:   ioDimensions,
				},
				model.Metric{
					Namespace:    "System",
					SubNamespace: "DiskIO",
					Name:         "diskio.read_time",
					Timestamp:    now,
					Value:        float64(ioStat.ReadTime),
					Unit:         "milliseconds",
					Dimensions:   ioDimensions,
				},
				model.Metric{
					Namespace:    "System",
					SubNamespace: "DiskIO",
					Name:         "diskio.write_time",
					Timestamp:    now,
					Value:        float64(ioStat.WriteTime),
					Unit:         "milliseconds",
					Dimensions:   ioDimensions,
				},
				model.Metric{
					Namespace:    "System",
					SubNamespace: "DiskIO",
					Name:         "diskio.io_time",
					Timestamp:    now,
					Value:        float64(ioStat.IoTime),
					Unit:         "milliseconds",
					Dimensions:   ioDimensions,
				},

				model.Metric{
					Namespace:    "System",
					SubNamespace: "DiskIO",
					Name:         "diskio.merged_read_count", // Number of reads merged
					Timestamp:    now,
					Value:        float64(ioStat.MergedReadCount),
					Unit:         "count",
					Dimensions:   ioDimensions, // Use dimensions possibly updated with SerialNumber
				},
				model.Metric{
					Namespace:    "System",
					SubNamespace: "DiskIO",
					Name:         "diskio.merged_write_count", // Number of writes merged
					Timestamp:    now,
					Value:        float64(ioStat.MergedWriteCount),
					Unit:         "count",
					Dimensions:   ioDimensions,
				},
				model.Metric{
					Namespace:    "System",
					SubNamespace: "DiskIO",
					Name:         "diskio.weighted_io", // Time spent doing I/Os (ms)
					Timestamp:    now,
					Value:        float64(ioStat.WeightedIO), // Removed as WeightedIOtime is not defined
					Unit:         "milliseconds",
					Dimensions:   ioDimensions,
				},
				// Add BusyTime if using gopsutil v3.3.0+
				// model.Metric{
				// 	Namespace:  "System/DiskIO",
				// 	Name:       "diskio.busy_time", // Time disk spent busy (Linux only, requires kernel 4.18+)
				// 	Timestamp:  now,
				// 	Value:      float64(ioStat.BusyTime),
				// 	Unit:       "milliseconds",
				// 	Dimensions: ioDimensions,
				// },
			)
		}
	}

	// Return all collected metrics and nil error (or the first critical error encountered)
	return metrics, nil
}
