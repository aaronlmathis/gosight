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
along with LeetScraper. If not, see https://www.gnu.org/licenses/.
*/

// gosight/agent/internal/collector/system/cpu.go
// Package system provides collectors for system hardware (CPU/RAM/DISK/ETC)

package system

import (
	"context"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/shirou/gopsutil/v3/cpu"
)

// Collector is the interface that all metric collectors must implement.
type CPUCollector struct{}

func (c *CPUCollector) Name() string {
	return "cpu"
}

func NewCPUCollector() *CPUCollector {
	return &CPUCollector{}
}

func (c *CPUCollector) Collect(ctx context.Context) ([]model.Metric, error) {
	percentages, err := cpu.PercentWithContext(ctx, 200*time.Millisecond, false)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now()
	metrics := make([]model.Metric, 0, len(percentages))
	for i, usage := range percentages {
		metrics = append(metrics, model.Metric{
			Namespace: "System/CPU",
			Name:      "cpu.total",
			Timestamp: timestamp,
			Value:     usage,
			Unit:      "percent",
			Dimensions: map[string]string{
				"core": "all",
			},
		})
		if i == 0 {
			break // since we're not using per-core values
		}
	}
	return metrics, nil
}
