package runner

import (
	"context"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/collector"
	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/agent/internal/meta"
	"github.com/aaronlmathis/gosight/agent/internal/sender"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// RunAgent starts the agent's collection loop and sends tasks to the pool
func RunAgent(ctx context.Context, cfg *config.Config) {
	reg := collector.NewRegistry(cfg)
	sndr, err := sender.NewSender(ctx, cfg)
	if err != nil {
		utils.Fatal("❌ Failed to connect to server: %v", err)
	}
	defer sndr.Close()

	taskQueue := make(chan model.MetricPayload, 500)
	go sender.StartWorkerPool(ctx, sndr, taskQueue, 10)

	ticker := time.NewTicker(cfg.Agent.Interval)
	defer ticker.Stop()

	utils.Info("🚀 Agent started. Sending metrics every %v", cfg.Agent.Interval)

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

			var hostMetrics []model.Metric
			containerBatches := make(map[string][]model.Metric)
			containerMetas := make(map[string]*model.Meta)

			for _, m := range metrics {
				if len(m.Dimensions) > 0 && m.Dimensions["container_id"] != "" {
					id := m.Dimensions["container_id"]
					if id == "" {
						continue
					}
					// Add container metrics to containerBatches
					containerBatches[id] = append(containerBatches[id], m)

					// Initialize Meta only once per container ID
					containerMeta, exists := containerMetas[id]
					if !exists {
						containerMeta = meta.BuildContainerMeta(cfg, nil)
						containerMetas[id] = containerMeta
					}

					//utils.Debug("🔍 Dimensions given with: %s - %v", id, m.Dimensions)
					// Populate meta with container-specific information
					for k, v := range m.Dimensions {
						switch k {
						case "container_id":
							containerMeta.ContainerID = v
							containerMeta.Tags["container_id"] = v
						case "name", "container_name":
							containerMeta.ContainerName = v
							containerMeta.Tags["container_name"] = v
						case "image":
							containerMeta.ImageID = v
							containerMeta.Tags["image"] = v
						default:
							containerMeta.Tags[k] = v
						}
					}
					meta.BuildStandardTags(containerMeta, m, true)
					//utils.Debug("🔍 Container Meta Tags: %v", containerMeta.Tags)
				} else {
					// Host metrics, collect them separately
					hostMetrics = append(hostMetrics, m)
				}
			}

			// Send host metrics as a single payload
			if len(hostMetrics) > 0 {
				hostMeta := meta.BuildHostMeta(cfg, nil)
				meta.BuildStandardTags(hostMeta, hostMetrics[0], false)

				payload := model.MetricPayload{
					Host:      cfg.Agent.HostOverride,
					Timestamp: time.Now(),
					Metrics:   hostMetrics,
					Meta:      hostMeta,
				}
				//utils.Info("📦 META Payload for: %s - %v", payload.Host, payload.Meta)
				select {
				case taskQueue <- payload:
				default:
					utils.Warn("⚠️ Host task queue full! Dropping host metrics")
				}
			}

			// Send each container as a separate payload
			for id, metrics := range containerBatches {
				payload := model.MetricPayload{
					Host:      cfg.Agent.HostOverride,
					Timestamp: time.Now(),
					Metrics:   metrics,
					Meta:      containerMetas[id],
				}
				utils.Info("📦 META Payload for: %s - %v", payload.Host, payload.Meta)

				select {
				case taskQueue <- payload:
				default:
					utils.Warn("⚠️ Task queue full! Dropping container metrics for %s", id)
				}
			}
		}
	}
}
