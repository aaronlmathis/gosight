package runner

import (
	"context"
	"os"
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

	taskQueue := make(chan model.MetricPayload, 500)
	go sender.StartWorkerPool(ctx, sndr, taskQueue, 10)

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
				if len(m.Dimensions) > 0 && m.Dimensions["container_id"] != "" {
					id := m.Dimensions["container_id"]
					if id == "" {
						continue
					}
					// Add container metrics to containerBatches
					containerBatches[id] = append(containerBatches[id], m)

					// Initialize and populate container meta if not already done
					meta, ok := containerMetas[id]
					if !ok {
						meta = &model.Meta{
							Tags: make(map[string]string),
						}
						containerMetas[id] = meta
					}

					// Populate meta with container-specific information
					for k, v := range m.Dimensions {
						switch k {
						case "container_id":
							meta.ContainerID = v
							meta.Tags["container_id"] = v
						case "container_name":
							meta.ContainerName = v
							meta.Tags["container_name"] = v
						case "image":
							meta.ImageID = v
							meta.Tags["image"] = v
						default:
							meta.Tags[k] = v
						}
					}

					// Populate meta with common tags like hostname and IP
					meta.Hostname = hostname
					meta.IPAddress = utils.GetLocalIP()

					// Generate endpoint ID for this container
					if meta.Tags == nil {
						meta.Tags = make(map[string]string)
					}
					meta.Tags["endpoint_id"] = utils.GenerateEndpointID(meta)
				} else {
					// Host metrics, collect them separately
					hostMetrics = append(hostMetrics, m)
				}
			}

			// Send host metrics as a single payload
			if len(hostMetrics) > 0 {
				meta := meta.BuildMeta(cfg, map[string]string{
					"job":      hostMetrics[0].Namespace,
					"instance": hostname,
				})
				endpointID := utils.GenerateEndpointID(meta)
				meta.Tags["instance"] = endpointID
				meta.Tags["endpoint_id"] = endpointID // Attach endpoint ID to host metrics
				payload := model.MetricPayload{
					Host:      cfg.HostOverride,
					Timestamp: time.Now(),
					Metrics:   hostMetrics,
					Meta:      meta,
				}
				//utils.Info("📦 META Payload for host. %s -  %s - %s", payload.Meta.Tags["endpoint_id"], payload.Meta.Tags["instance"], payload.Meta.Tags["job"])
				select {
				case taskQueue <- payload:
				default:
					utils.Warn("⚠️ Host task queue full! Dropping host metrics")
				}
			}

			// Send each container as a separate payload
			for id, metrics := range containerBatches {
				payload := model.MetricPayload{
					Host:      cfg.HostOverride,
					Timestamp: time.Now(),
					Metrics:   metrics,
					Meta:      containerMetas[id],
				}
				//utils.Info("📦 META Payload for host. %s -  %s - %s", payload.Meta.Tags["endpoint_id"], payload.Meta.Tags["instance"], payload.Meta.Tags["job"])

				select {
				case taskQueue <- payload:
				default:
					utils.Warn("⚠️ Task queue full! Dropping container metrics for %s", id)
				}
			}
		}
	}
}
