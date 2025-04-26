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

// gosight/server/internal/api
// metrics.go - gRPC handler for metrics submission

package telemetry

import (
	"github.com/aaronlmathis/gosight/server/internal/sys"
	"github.com/aaronlmathis/gosight/shared/model"
	pb "github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// MetricsHandler implements pb.MetricsServiceServer
// MetricsHandler implements MetricsServiceServer
type MetricsHandler struct {
	Sys *sys.SystemContext
	pb.UnimplementedMetricsServiceServer
}

func NewMetricsHandler(sys *sys.SystemContext) *MetricsHandler {
	utils.Debug("MetricsHandler initialized with store: %T", sys.Stores.Metrics)
	return &MetricsHandler{
		Sys: sys,
	}
}

func (h *MetricsHandler) SubmitStream(stream pb.MetricsService_SubmitStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				return stream.SendAndClose(&pb.MetricResponse{
					Status:     "ok",
					StatusCode: 0,
				})
			}
			utils.Error("Stream receive error: %v", err)
			return err
		}

		// Convert payload into a model.MetricPayload.
		converted := ConvertToModelPayload(req)

		//utils.Debug("Received metrics from %s", converted.Meta.EndpointID)

		// Check rules
		h.Sys.Tele.Evaluator.Evaluate(h.Sys.Ctx, converted.Metrics, converted.Meta)

		// Update Agent Tracker
		h.Sys.Tracker.UpdateAgent(converted.Meta)
		if converted.Meta.ContainerID != "" {
			h.Sys.Tracker.UpdateContainer(converted.Meta)
		}
		// Broadcast to WebSocket clients
		h.Sys.Web.BroadcastMetric(converted)

		//utils.Debug("Received metrics: %v", converted)

		// Enqueue metrics for storage
		if err := h.Sys.Stores.Metrics.Write([]model.MetricPayload{converted}); err != nil {
			utils.Warn("Failed to enqueue metrics from %s: %v", converted.EndpointID, err)
		} else {
			utils.Info("Enqueued %d metrics from host: %s at %s", len(converted.Metrics), converted.EndpointID, converted.Timestamp)

			if converted.Meta != nil && converted.Meta.EndpointID != "" {
				// Store the meta information in the MetaTracker
				h.Sys.Tele.Meta.Set(converted.Meta.EndpointID, *converted.Meta)
			} else {
				utils.Debug("Missing EndpointID — not storing meta")
			}

			for _, m := range converted.Metrics {
				// Index the metric in the MetricIndex
				h.Sys.Tele.Index.Add(m.Namespace, m.SubNamespace, m.Name, m.Dimensions)

			}
		}
	}
}
