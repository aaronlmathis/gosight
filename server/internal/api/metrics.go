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

package api

import (
	"context"
	"log"

	"github.com/aaronlmathis/gosight/shared/proto"
)

// MetricsHandler implements pb.MetricsServiceServer
// MetricsHandler implements MetricsServiceServer
type MetricsHandler struct {
	proto.UnimplementedMetricsServiceServer
}

func (h *MetricsHandler) SubmitMetrics(ctx context.Context, req *proto.MetricPayload) (*proto.MetricResponse, error) {
	converted := ConvertToModelPayload(req)

	log.Printf("Received metrics from host: %s at %s", converted.Host, converted.Timestamp)
	for _, m := range converted.Metrics {
		log.Printf(" - %s: %.2f %s", m.Name, m.Value, m.Unit)
	}

	// TODO: Store the converted metrics in DB or processing pipeline

	return &proto.MetricResponse{
		Status:     "ok",
		StatusCode: 0,
	}, nil
}
