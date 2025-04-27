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
	"math"
	"time"

	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
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

	// --- Virtual Memory ---
	memory, err := mem.VirtualMemory()
	if err != nil {
		utils.Warn("Error getting memory info: %v", err)
	} else if memory != nil {
		dims := map[string]string{"source": "physical"}

		metrics = append(metrics,
			agentutils.Metric("System", "Memory", "total", memory.Total, "gauge", "bytes", dims, now),
			agentutils.Metric("System", "Memory", "available", memory.Available, "gauge", "bytes", dims, now),
			agentutils.Metric("System", "Memory", "used", memory.Used, "gauge", "bytes", dims, now),
			agentutils.Metric("System", "Memory", "used_percent", memory.UsedPercent, "gauge", "percent", dims, now),
		)
	}

	// --- Swap Memory ---
	swap, err := mem.SwapMemory()
	if err != nil {
		utils.Warn("Error getting swap memory info: %v", err)
	} else if swap != nil && swap.Total > 0 {
		dims := map[string]string{"source": "swap"}
		utils.Debug("🧠 swap.Total = %d, swap.Free = %d", swap.Total, swap.Free)
		usedPercent := swap.UsedPercent
		if usedPercent <= 0 {
			usedPercent = float64(swap.Total-swap.Free) / float64(swap.Total) * 100
		}

		if math.IsNaN(usedPercent) || math.IsInf(usedPercent, 0) {
			utils.Warn("🚨 Swap used_percent computed invalid value: %v", usedPercent)
			usedPercent = 0
		}

		utils.Debug("🧠 swap_used_percent = %.2f", usedPercent)

		metrics = append(metrics,
			agentutils.Metric("System", "Memory", "swap_total", swap.Total, "gauge", "bytes", dims, now),
			agentutils.Metric("System", "Memory", "swap_used", swap.Used, "gauge", "bytes", dims, now),
			agentutils.Metric("System", "Memory", "swap_free", swap.Free, "gauge", "bytes", dims, now),
			agentutils.Metric("System", "Memory", "swap_used_percent", usedPercent, "gauge", "percent", dims, now),
		)
	} else {
		utils.Debug("🚫 Swap metrics skipped — no swap memory available.")
	}

	return metrics, nil
}
