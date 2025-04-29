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

	"github.com/aaronlmathis/gosight/agent/internal/command"
	"github.com/aaronlmathis/gosight/agent/internal/config"
	grpcconn "github.com/aaronlmathis/gosight/agent/internal/grpc"
	"github.com/aaronlmathis/gosight/agent/internal/protohelper"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
	"google.golang.org/grpc"
	goproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Sender holds the gRPC client and connection
type MetricSender struct {
	client proto.StreamServiceClient
	cc     *grpc.ClientConn
	stream proto.StreamService_StreamClient
	wg     sync.WaitGroup
}

// NewSender establishes a gRPC connection
func NewSender(ctx context.Context, cfg *config.Config) (*MetricSender, error) {

	clientConn, err := grpcconn.GetGRPCConn(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Create gRPC client
	// and establish a stream for sending metrics
	client := proto.NewStreamServiceClient(clientConn)
	utils.Info("established gRPC Connection with %v", cfg.Agent.ServerURL)

	//
	stream, err := client.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	sender := &MetricSender{
		client: client,
		cc:     clientConn,
		stream: stream,
	}
	go sender.receiveResponses()

	return sender, nil
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

	metricPayload := &proto.MetricPayload{
		AgentId:    payload.AgentID,
		HostId:     payload.HostID,
		Hostname:   payload.Hostname,
		EndpointId: payload.EndpointID,
		Timestamp:  timestamppb.New(payload.Timestamp),
		Metrics:    pbMetrics,
		Meta:       convertedMeta,
	}

	b, err := goproto.Marshal(metricPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal MetricPayload: %w", err)
	}

	streamPayload := &proto.StreamPayload{
		Payload: &proto.StreamPayload_Metric{
			Metric: &proto.MetricWrapper{
				RawPayload: b,
			},
		},
	}
	// Send StreamPayload now
	if err := s.stream.Send(streamPayload); err != nil {
		return fmt.Errorf("stream send failed: %w", err)
	}

	//utils.Info("Streamed %d metrics", len(pbMetrics))
	return nil
}

// receiveResponses listens for commands sent from server.

func (s *MetricSender) receiveResponses() {
	for {
		resp, err := s.stream.Recv()
		if err != nil {
			utils.Error("Stream receive error: %v", err)
			break // Exit loop (can reconnect later if you want)
		}

		utils.Debug("Received StreamResponse: status=%s", resp.Status)

		if resp.Command != nil {
			utils.Info("Received CommandRequest: type=%s command=%s", resp.Command.CommandType, resp.Command.Command)

			// Call command handler
			command.HandleCommand(resp.Command)
		}
	}
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
