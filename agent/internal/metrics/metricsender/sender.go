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
	"fmt"
	"sync"
	"time"

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
	cfg    *config.Config
	ctx    context.Context
}

// NewSender establishes a gRPC connection
// and creates a stream for sending metrics
// It returns a MetricSender instance
func NewSender(ctx context.Context, cfg *config.Config) (*MetricSender, error) {

	clientConn, err := grpcconn.GetGRPCConn(cfg)
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
		ctx:    ctx,
		client: client,
		cc:     clientConn,
		stream: stream,
		cfg:    cfg,
	}
	go sender.receiveResponses()

	return sender, nil
}

// Close waits for all workers to finish and closes the gRPC connection
// It returns an error if the connection could not be closed
// or if any worker failed
func (s *MetricSender) Close() error {
	utils.Info("Closing MetricSender... waiting for workers")
	s.wg.Wait()
	utils.Info("All MetricSender workers finished")
	return s.cc.Close()
}

// SendMetrics sends a MetricPayload to the server
// It converts the MetricPayload to a protobuf message
// and sends it over the gRPC stream
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
	sendCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sendCh := make(chan error, 1)
	go func() {
		sendCh <- s.stream.Send(streamPayload)
	}()

	select {
	case err := <-sendCh:
		if err != nil {
			return fmt.Errorf("stream send failed: %w", err)
		}
		return nil
	case <-sendCtx.Done():
		return fmt.Errorf("stream send timeout")
	}
}

// receiveResponses listens for commands sent from server.
// It handles the commands and sends back responses.
// It runs in a separate goroutine and handles reconnections
// in case of errors.
func (s *MetricSender) receiveResponses() {
	for {
		resp, err := s.stream.Recv()
		if err != nil {
			select {
			case <-s.ctx.Done():
				utils.Info("Context canceled, exiting receive loop cleanly")
				return
			default:
				utils.Error("Stream receive error: %v", err)

				if recErr := s.reconnectStream(); recErr != nil {
					utils.Error("Failed to reconnect stream: %v", recErr)
					return
				}
				utils.Info("Successfully reconnected stream after failure")
				continue
			}
		}

		utils.Debug("Received StreamResponse: status=%s", resp.Status)
		utils.Debug("Response Payload: %v", resp)
		if resp.Command != nil {
			utils.Info("Received CommandRequest: type=%s command=%s", resp.Command.CommandType, resp.Command.Command)
			result := command.HandleCommand(s.ctx, resp.Command)
			if result != nil {
				s.sendCommandResponseWithRetry(result)
			}
		}
	}
}

// reconnectStream attempts to reconnect the gRPC stream.
// It closes the old connection and creates a new one.
// It returns an error if the reconnection fails.
func (s *MetricSender) reconnectStream() error {
	var err error
	// Close old connection safely if you want (optional)
	if s.cc != nil {
		_ = s.cc.Close()
	}

	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	// Rebuild gRPC client connection
	conn, err := grpcconn.GetGRPCConn(s.cfg) // Use your same logic
	if err != nil {
		return fmt.Errorf("failed to reconnect gRPC: %w", err)
	}

	s.cc = conn
	s.client = proto.NewStreamServiceClient(conn)

	stream, err := s.client.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to reopen stream: %w", err)
	}

	s.stream = stream
	return nil
}

// sendCommandResponseWithRetry attempts to send a CommandResponse with retries
// It uses exponential backoff for retries
// It sends the CommandResponse over the gRPC stream
// It handles errors and timeouts
func (s *MetricSender) sendCommandResponseWithRetry(resp *proto.CommandResponse) {
	const maxAttempts = 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		sendCtx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- s.stream.Send(&proto.StreamPayload{
				Payload: &proto.StreamPayload_CommandResponse{
					CommandResponse: resp,
				},
			})
		}()

		select {
		case err := <-done:
			if err != nil {
				utils.Warn("Send attempt %d failed: %v", attempt, err)
				_ = s.reconnectStream()
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			utils.Debug("CommandResponse sent on attempt %d", attempt)
			return
		case <-sendCtx.Done():
			utils.Warn("Send attempt %d timed out", attempt)
			_ = s.reconnectStream()
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	utils.Error("Failed to send CommandResponse after %d attempts: %v", maxAttempts, resp)
}
