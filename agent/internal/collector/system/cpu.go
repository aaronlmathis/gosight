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

// gosight/agent/internal/collector/system/cpu.go
// Package system provides collectors for system hardware (CPU/RAM/DISK/ETC)
// cpu.go collects metrics on cpu usage, times, and info.
// It uses the gopsutil library to gather CPU metrics.

package system

import (
	"context"
	"strconv"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/shirou/gopsutil/v4/cpu"
)

type CPUCollector struct{}

func NewCPUCollector() *CPUCollector {
	return &CPUCollector{}
}

func (c *CPUCollector) Name() string {
	return "cpu"
}

func (c *CPUCollector) Collect(ctx context.Context) ([]model.Metric, error) {
	var metrics []model.Metric
	now := time.Now()

	// Per-core usage
	if percentPerCore, err := cpu.PercentWithContext(ctx, 200*time.Millisecond, true); err == nil {
		for i, val := range percentPerCore {
			metrics = append(metrics, model.Metric{
				Namespace:    "System",
				SubNamespace: "CPU",
				Name:         "usage_percent",
				Timestamp:    now,
				Value:        val,
				Type:         "gauge",
				Unit:         "percent",
				Dimensions: map[string]string{
					"core":  formatCore(i),
					"scope": "per_core",
				},
			})
		}
	}

	// Total CPU usage
	if percentTotal, err := cpu.PercentWithContext(ctx, 200*time.Millisecond, false); err == nil && len(percentTotal) > 0 {
		metrics = append(metrics, model.Metric{
			Namespace:    "System",
			SubNamespace: "CPU",
			Name:         "usage_percent",
			Timestamp:    now,
			Value:        percentTotal[0],
			Type:         "gauge",
			Unit:         "percent",
			Dimensions: map[string]string{
				"scope": "total",
			},
		})
	}

	// CPU times (cumulative)
	if times, err := cpu.TimesWithContext(ctx, false); err == nil && len(times) > 0 {
		t := times[0]
		for k, v := range map[string]float64{
			"user":       t.User,
			"system":     t.System,
			"idle":       t.Idle,
			"nice":       t.Nice,
			"iowait":     t.Iowait,
			"irq":        t.Irq,
			"softirq":    t.Softirq,
			"steal":      t.Steal,
			"guest":      t.Guest,
			"guest_nice": t.GuestNice,
		} {
			metrics = append(metrics, model.Metric{
				Namespace:    "System",
				SubNamespace: "CPU",
				Name:         "time_" + k,
				Timestamp:    now,
				Value:        v,
				Type:         "counter",
				Unit:         "seconds",
				Dimensions: map[string]string{
					"scope": "total",
				},
			})
		}
	}

	// CPU Info: Clock speed per core
	if info, err := cpu.InfoWithContext(ctx); err == nil && len(info) > 0 {
		for i, cpu := range info {
			metrics = append(metrics, model.Metric{
				Namespace:    "System",
				SubNamespace: "CPU",
				Name:         "clock_mhz",
				Timestamp:    now,
				Value:        cpu.Mhz,
				Type:         "gauge",
				Unit:         "MHz",
				Dimensions: map[string]string{
					"core":     formatCore(i),
					"vendor":   cpu.VendorID,
					"model":    cpu.ModelName,
					"family":   cpu.Family,
					"physical": formatBool(cpu.PhysicalID != ""),
				},
			})
		}
	}

	// Logical and physical core counts
	if count, err := cpu.CountsWithContext(ctx, true); err == nil {
		metrics = append(metrics, model.Metric{
			Namespace:    "System",
			SubNamespace: "CPU",
			Name:         "count_logical",
			Timestamp:    now,
			Value:        float64(count),
			Type:         "gauge",
			Unit:         "count",
		})
	}
	if count, err := cpu.CountsWithContext(ctx, false); err == nil {
		metrics = append(metrics, model.Metric{
			Namespace:    "System",
			SubNamespace: "CPU",
			Name:         "count_physical",
			Timestamp:    now,
			Value:        float64(count),
			Type:         "gauge",
			Unit:         "count",
		})
	}

	for _, m := range metrics {
		utils.Debug("📡 Collector output: %+v", m)
	}
	return metrics, nil
}

func formatCore(i int) string {
	return "core" + strconv.Itoa(i)
}

func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
