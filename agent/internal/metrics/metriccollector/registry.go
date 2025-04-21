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

// gosight/agent/internal/collector/registry.go
// registry.go - loads and initializes all enabled collectors at runtime.

package metriccollector

import (
	"context"
	"log"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/agent/internal/metrics/metriccollector/container"
	"github.com/aaronlmathis/gosight/agent/internal/metrics/metriccollector/system"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// Registry holds active collectors keyed by name
type MetricRegistry struct {
	Collectors map[string]MetricCollector
}

// NewRegistry initializes and registers enabled collectors
func NewRegistry(cfg *config.Config) *MetricRegistry {
	reg := &MetricRegistry{Collectors: make(map[string]MetricCollector)}
	log.Printf("🔍 Available collectors: %v", func() []string {
		collectors := []string{}
		for _, name := range cfg.Agent.MetricsEnabled {
			collectors = append(collectors, name)
		}
		return collectors
	})
	for _, name := range cfg.Agent.MetricsEnabled {
		switch name {
		case "cpu":
			reg.Collectors["cpu"] = system.NewCPUCollector()
		case "mem":
			reg.Collectors["mem"] = system.NewMemCollector()
		case "disk":
			reg.Collectors["disk"] = system.NewDiskCollector()
		case "host":
			reg.Collectors["host"] = system.NewHostCollector()
		case "net":
			reg.Collectors["net"] = system.NewNetworkCollector()
		case "podman":
			reg.Collectors["podman"] = container.NewPodmanCollectorWithSocket(cfg.Podman.Socket)
		case "docker":
			reg.Collectors["docker"] = container.NewDockerCollectorWithSocket(cfg.Docker.Socket)
		default:
			utils.Warn("⚠️ Unknown collector: %s (skipping) \n", name)
		}
	}
	utils.Info("Loaded %d metric collectors", len(reg.Collectors))

	return reg
}

// Collect runs all active collectors and returns all collected metrics
func (r *MetricRegistry) Collect(ctx context.Context) ([]model.Metric, error) {
	var all []model.Metric

	for name, collector := range r.Collectors {
		metrics, err := collector.Collect(ctx)
		if err != nil {
			utils.Error("❌ Error collecting %s: %v\n", name, err)
			continue
		}
		all = append(all, metrics...)
	}

	return all, nil
}
