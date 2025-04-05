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

package api

import (
	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/shared/model"
	pb "github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// MetricsHandler implements pb.MetricsServiceServer
// MetricsHandler implements MetricsServiceServer
type MetricsHandler struct {
	store       store.MetricStore
	Tracker     *store.AgentTracker
	metricIndex *store.MetricIndex
	pb.UnimplementedMetricsServiceServer
}

func NewMetricsHandler(s store.MetricStore, tracker *store.AgentTracker, metricIndex *store.MetricIndex) *MetricsHandler {
	utils.Debug("🚀 MetricsHandler initialized with store: %T", s)
	return &MetricsHandler{
		store:       s,
		Tracker:     tracker,
		metricIndex: metricIndex,
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
			utils.Error("❌ Stream receive error: %v", err)
			return err
		}

		converted := ConvertToModelPayload(req)
		utils.Debug("📦 Received metrics from host: %s at %s", converted.Metrics[0].Namespace, converted.Metrics[0].SubNamespace)
		if err := h.store.Write([]model.MetricPayload{converted}); err != nil {
			utils.Warn("❌ Failed to enqueue metrics from %s: %v", converted.Host, err)
		} else {
			utils.Info("📥 Enqueued %d metrics from host: %s at %s", len(converted.Metrics), converted.Host, converted.Timestamp)

			if converted.Meta.Hostname != "" {
				h.Tracker.UpdateAgent(*converted.Meta)
			}
			for _, m := range converted.Metrics {
				h.metricIndex.Add(m.Namespace, m.SubNamespace, m.Name, m.Dimensions)
				utils.Debug("🧩 Indexed: %s / %s / %s", m.Namespace, m.SubNamespace, m.Name)
			}
		}
	}
}
