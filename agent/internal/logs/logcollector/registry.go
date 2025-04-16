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

// gosight/agent/internal/logs/logcollector/registry.go
// registry.go - loads and initializes all enabled log collectors at runtime.

package logcollector

import (
	"context"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	linuxcollector "github.com/aaronlmathis/gosight/agent/internal/logs/logcollector/linux"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type LogRegistry struct {
	LogCollectors map[string]Collector
}

// NewRegistry initializes and registers enabled log collectors
func NewRegistry(cfg *config.Config) *LogRegistry {
	reg := &LogRegistry{LogCollectors: make(map[string]Collector)}

	for _, name := range cfg.Agent.LogCollection.Sources {
		switch name {
		case "journald":
			reg.LogCollectors["journald"] = linuxcollector.NewJournaldCollector(cfg)

		default:
			utils.Warn("⚠️ Unknown collector: %s (skipping) \n", name)
		}
	}
	utils.Info("Loaded %d log collectors", len(reg.LogCollectors))

	return reg
}

// Collect runs all active collectors and returns all collected metrics
func (r *LogRegistry) Collect(ctx context.Context) ([][]model.LogEntry, error) {
	var allBatches [][]model.LogEntry

	for name, collector := range r.LogCollectors {
		logBatches, err := collector.Collect(ctx)
		if err != nil {
			utils.Error("Error collecting %s: %v\n", name, err)
			continue
		}
		allBatches = append(allBatches, logBatches...)
		utils.Debug("✔️ LogRegistry returned %d batches", len(logBatches))
	}

	return allBatches, nil
}
