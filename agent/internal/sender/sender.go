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

// gosight/agent/internal/sender/sender.go

package sender

import (
	"context"
	"fmt"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/agent/internal/utils"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Sender holds the gRPC client and connection
type Sender struct {
	client proto.MetricsServiceClient
	cc     *grpc.ClientConn
	stream proto.MetricsService_SubmitStreamClient
}

// NewSender establishes a gRPC connection
func NewSender(ctx context.Context, cfg *config.AgentConfig) (*Sender, error) {
	clientConn, err := grpc.NewClient(cfg.ServerURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	client := proto.NewMetricsServiceClient(clientConn)
	utils.Info("📤 established gRPC Connection with %v", cfg.ServerURL)

	stream, err := client.SubmitStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	return &Sender{
		client: client,
		cc:     clientConn,
		stream: stream,
	}, nil

}

// Close the gRPC connection
func (s *Sender) Close() error {
	return s.cc.Close()
}

func (s *Sender) SendMetrics(payload model.MetricPayload) error {
	pbMetrics := make([]*proto.Metric, 0, len(payload.Metrics))
	for _, m := range payload.Metrics {
		pbMetric := &proto.Metric{
			Name:      m.Name,
			Value:     m.Value,
			Unit:      m.Unit,
			Timestamp: timestamppb.New(m.Timestamp),
		}
		pbMetrics = append(pbMetrics, pbMetric)
	}

	req := &proto.MetricPayload{
		Host:      payload.Host,
		Timestamp: timestamppb.New(payload.Timestamp),
		Metrics:   pbMetrics,
		Meta:      payload.Meta,
	}

	if err := s.stream.Send(req); err != nil {
		return fmt.Errorf("stream send failed: %w", err)
	}

	utils.Info("📤 Streamed %d metrics", len(pbMetrics))
	return nil
}
