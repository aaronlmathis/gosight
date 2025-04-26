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

// gosight/agent/internal/sender/sender.go

package metricsender

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/agent/internal/protohelper"
	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Sender holds the gRPC client and connection
type MetricSender struct {
	client proto.MetricsServiceClient
	cc     *grpc.ClientConn
	stream proto.MetricsService_SubmitStreamClient
	wg     sync.WaitGroup
}

// NewSender establishes a gRPC connection
func NewSender(ctx context.Context, cfg *config.Config) (*MetricSender, error) {

	// Load TLS config for agent
	tlsCfg, err := agentutils.LoadTLSConfig(cfg)
	if err != nil {
		return nil, err
	}

	// add mTLS to degug log.
	if len(tlsCfg.Certificates) > 0 {
		utils.Info("using mTLS for agent authentication")
	} else {
		utils.Info("Using TLS only (no client certificate)")
	}

	// Establish gRPC connection
	clientConn, err := grpc.NewClient(cfg.Agent.ServerURL,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)),
	)
	if err != nil {
		return nil, err
	}
	utils.Info("connecting to server at: %s", cfg.Agent.ServerURL)
	// Create gRPC client
	// and establish a stream for sending metrics
	client := proto.NewMetricsServiceClient(clientConn)
	utils.Info("established gRPC Connection with %v", cfg.Agent.ServerURL)

	//
	stream, err := client.SubmitStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	return &MetricSender{
		client: client,
		cc:     clientConn,
		stream: stream,
	}, nil

}

func (s *MetricSender) Close() error {
	utils.Info("Closing MetricSender... waiting for workers")
	s.wg.Wait()
	utils.Info("All MetricSender workers finished")
	return s.cc.Close()
}

func (s *MetricSender) SendMetrics(payload *model.MetricPayload) error {
	pbMetrics := make([]*proto.Metric, 0, len(payload.Metrics))
	for _, m := range payload.Metrics {
		pbMetric := &proto.Metric{
			Name:         m.Name,
			Namespace:    m.Namespace,
			Subnamespace: m.SubNamespace,
			Timestamp:    timestamppb.New(m.Timestamp),

			Value:             m.Value,
			Unit:              m.Unit,
			StorageResolution: int32(m.StorageResolution),
			Type:              m.Type,
			Dimensions:        m.Dimensions,
		}
		if m.StatisticValues != nil {
			pbMetric.StatisticValues = &proto.StatisticValues{
				Minimum:     m.StatisticValues.Minimum,
				Maximum:     m.StatisticValues.Maximum,
				SampleCount: int32(m.StatisticValues.SampleCount),
				Sum:         m.StatisticValues.Sum,
			}
		}
		pbMetrics = append(pbMetrics, pbMetric)
	}
	var convertedMeta *proto.Meta

	// Convert meta to proto
	if payload.Meta != nil {
		convertedMeta = protohelper.ConvertMetaToProtoMeta(payload.Meta)
	}
	//utils.Debug("Proto Meta Tags: %+v", convertedMeta)

	req := &proto.MetricPayload{
		AgentId:  payload.AgentID,
		HostId:   payload.HostID,
		Hostname: payload.Hostname,

		EndpointId: payload.EndpointID,
		Timestamp:  timestamppb.New(payload.Timestamp),
		Metrics:    pbMetrics,
		Meta:       convertedMeta,
	}
	//fmt.Printf("Sending proto.Meta: %+v\n", req.Meta)
	//utils.Debug("Sending %d metrics to server: %v", len(pbMetrics), pbMetrics)
	if err := s.stream.Send(req); err != nil {
		return fmt.Errorf("stream send failed: %w", err)
	}

	//utils.Info("Streamed %d metrics", len(pbMetrics))
	return nil
}

func AppendMetricsToFile(payload *model.MetricPayload, filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = file.Write(append(data, '\n')) // newline-delimited JSON
	return err
}
