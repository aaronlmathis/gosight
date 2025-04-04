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

// gosight/agent/internal/runner/runner.go

package runner

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/collector"
	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/agent/internal/meta"
	"github.com/aaronlmathis/gosight/agent/internal/sender"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// RunAgent starts the agent's collection loop and sends tasks to the pool
func RunAgent(ctx context.Context, cfg *config.AgentConfig) {
	reg := collector.NewRegistry(cfg)
	sndr, err := sender.NewSender(ctx, cfg)
	if err != nil {
		utils.Fatal("❌ Failed to connect to server: %v", err)
	}
	defer sndr.Close()

	taskQueue := make(chan model.MetricPayload, 100)
	go sender.StartWorkerPool(ctx, sndr, taskQueue, 5)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	utils.Info("🚀 Agent started. Sending metrics every %v", cfg.Interval)

	for {
		select {
		case <-ctx.Done():
			utils.Warn("🔌 Agent shutting down...")
			return
		case <-ticker.C:
			metrics, err := reg.Collect(ctx)
			if err != nil {
				utils.Error("❌ Metric collection failed: %v", err)
				continue
			}

			hostname, err := os.Hostname()
			if err != nil {
				hostname = "unknown"
				utils.Warn("⚠️ Failed to get hostname: %v", err)
			}

			var hostMetrics []model.Metric
			containerBatches := make(map[string][]model.Metric)
			containerMetas := make(map[string]*model.Meta)

			for _, m := range metrics {
				if strings.HasPrefix(m.Name, "container.") {
					id := m.Dimensions["container_id"]
					if id == "" {
						continue
					}
					containerBatches[id] = append(containerBatches[id], m)

					meta := &model.Meta{
						Tags: make(map[string]string),
					}
					for k, v := range m.Dimensions {
						switch k {
						case "container_id":
							meta.ContainerID = v
							meta.Tags["container_id"] = v
						case "name":
							meta.ContainerName = v
							meta.Tags["container_name"] = v
						case "image":
							meta.ImageID = v
							meta.Tags["image"] = v
						default:
							meta.Tags[k] = v
						}
					}
					meta.Hostname = hostname
					meta.IPAddress = utils.GetLocalIP()
					meta.OS = "linux"
					containerMetas[id] = meta
				} else {
					hostMetrics = append(hostMetrics, m)
				}
			}

			// send host metrics as one payload
			if len(hostMetrics) > 0 {
				meta := meta.BuildMeta(cfg, map[string]string{
					"job":      "gosight-agent",
					"instance": hostname,
				})
				payload := model.MetricPayload{
					Host:      cfg.HostOverride,
					Timestamp: time.Now(),
					Metrics:   hostMetrics,
					Meta:      meta,
				}
				select {
				case taskQueue <- payload:
				default:
					utils.Warn("⚠️ Host task queue full! Dropping host metrics")
				}
			}

			// send each container as separate payload
			for id, metrics := range containerBatches {
				payload := model.MetricPayload{
					Host:      cfg.HostOverride,
					Timestamp: time.Now(),
					Metrics:   metrics,
					Meta:      containerMetas[id],
				}
				select {
				case taskQueue <- payload:
				default:
					utils.Warn("⚠️ Task queue full! Dropping container metrics for %s", id)
				}
			}
		}
	}
}
