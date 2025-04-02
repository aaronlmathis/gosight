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

// gosight/agent/internal/collector/system/mem.go
// Package system provides collectors for system hardware (CPU/RAM/DISK/ETC)
// memo.go collects metrics on memory usage and info.
// It uses the gopsutil library to gather CPU metrics.

package system

import (
	"context"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/shirou/gopsutil/v4/mem"
)

type MEMCollector struct{}

func NewMemCollector() *MEMCollector {
	return &MEMCollector{}
}

func (c *MEMCollector) Name() string {
	return "mem"
}

func (c *MEMCollector) Collect(ctx context.Context) ([]model.Metric, error) {
	var metrics []model.Metric

	now := time.Now()
	memory, err := mem.VirtualMemory()
	if err != nil {
		utils.Info("Error getting memory info: %v", err)
		utils.Debug("Error getting memory info: %v", err)
	} else if memory == nil {
		utils.Info("Memory info is nil")
		utils.Debug("Memory info is nil")
	} else {
		metrics = append(metrics, model.Metric{
			Namespace:  "System/Memory",
			Name:       "mem.total",
			Timestamp:  now,
			Value:      float64(memory.Total),
			Unit:       "bytes",
			Dimensions: map[string]string{"source": "physical"},
		})
		metrics = append(metrics, model.Metric{
			Namespace:  "System/Memory",
			Name:       "mem.available",
			Timestamp:  now,
			Value:      float64(memory.Available),
			Unit:       "bytes",
			Dimensions: map[string]string{"source": "physical"},
		})
		metrics = append(metrics, model.Metric{
			Namespace:  "System/Memory",
			Name:       "mem.used",
			Timestamp:  now,
			Value:      float64(memory.Used),
			Unit:       "bytes",
			Dimensions: map[string]string{"source": "physical"},
		})
		metrics = append(metrics, model.Metric{
			Namespace:  "System/Memory",
			Name:       "mem.used_percent",
			Timestamp:  now,
			Value:      memory.UsedPercent,
			Unit:       "percent",
			Dimensions: map[string]string{"source": "physical"},
		})
	}

	// Swap memory
	swap, err := mem.SwapMemory()
	if err != nil {
		utils.Info("Error getting swap memory info: %v", err)
		utils.Debug("Error getting swap memory info: %v", err)
	}
	if swap == nil {
		utils.Info("Swap memory info is nil")
		utils.Debug("Swap memory info is nil")
	} else {
		metrics = append(metrics, model.Metric{
			Namespace:  "System/Memory",
			Name:       "swap.total",
			Timestamp:  now,
			Value:      float64(swap.Total),
			Unit:       "bytes",
			Dimensions: map[string]string{"source": "swap"},
		})
		metrics = append(metrics, model.Metric{
			Namespace:  "System/Memory",
			Name:       "swap.used",
			Timestamp:  now,
			Value:      float64(swap.Used),
			Unit:       "bytes",
			Dimensions: map[string]string{"source": "swap"},
		})
		metrics = append(metrics, model.Metric{
			Namespace:  "System/Memory",
			Name:       "swap.available",
			Timestamp:  now,
			Value:      float64(swap.Free),
			Unit:       "bytes",
			Dimensions: map[string]string{"source": "swap"},
		})
		metrics = append(metrics, model.Metric{
			Namespace:  "System/Memory",
			Name:       "swap.used_percent",
			Timestamp:  now,
			Value:      swap.UsedPercent,
			Unit:       "percent",
			Dimensions: map[string]string{"source": "swap"},
		})
	}

	return metrics, nil
}
