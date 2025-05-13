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

package logsender

import (
	"context"
	"fmt"
	"sync"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	grpcconn "github.com/aaronlmathis/gosight/agent/internal/grpc"
	"github.com/aaronlmathis/gosight/agent/internal/protohelper"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	goproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Sender holds the gRPC client and connection
type LogSender struct {
	client proto.LogServiceClient
	cc     *grpc.ClientConn
	stream proto.LogService_SubmitStreamClient
	wg     sync.WaitGroup
	Config *config.Config
}

// NewSender establishes a gRPC connection to the server and creates a new LogSender instance.
// It takes a context and a configuration object as parameters.
// The function returns a pointer to the LogSender instance and an error if any occurs during the process.
// It also sets up a stream for sending log data to the server.
func NewSender(ctx context.Context, cfg *config.Config) (*LogSender, error) {
	clientConn, err := grpcconn.GetGRPCConn(cfg)
	if err != nil {
		return nil, err
	}

	client := proto.NewLogServiceClient(clientConn)
	utils.Info("established gRPC Connection with %v", cfg.Agent.ServerURL)

	//
	stream, err := client.SubmitStream(ctx)
	if err != nil {
		utils.Debug("Failed to open stream: %v", err)
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	return &LogSender{
		client: client,
		cc:     clientConn,
		stream: stream,
		Config: cfg,
	}, nil

}

// Close closes the gRPC connection and waits for all workers to finish.
// It returns an error if any occurs during the process.
// The function uses a wait group to ensure that all workers have completed their tasks
// before closing the connection. It also logs the status of the workers.
func (s *LogSender) Close() error {
	utils.Info("Closing LogSender... waiting for workers")
	s.wg.Wait()
	utils.Info("All LogSender workers finished")
	return s.cc.Close()
}

// SendLogs sends a log payload to the gRPC server.
// It takes a pointer to a LogPayload object as a parameter.
// The function converts the log payload to a protobuf format and sends it to the server.
func (s *LogSender) SendLogs(payload *model.LogPayload) error {
	pbLogs := make([]*proto.LogEntry, 0, len(payload.Logs))
	utils.Debug("Log Meta: %v", payload.Meta)
	for _, log := range payload.Logs {
		pbLog := &proto.LogEntry{
			Timestamp: timestamppb.New(log.Timestamp),
			Level:     log.Level,
			Message:   log.Message,
			Source:    log.Source,
			Category:  log.Category,
			Pid:       int32(log.PID),
			Fields:    log.Fields,
			Tags:      log.Tags,
			Meta:      protohelper.ConvertLogMetaToProto(log.Meta),
		}
		//utils.Debug("Sender: LogEntry: %v", pbLog)
		//utils.Debug("Sender:LogMeta: %v", pbLog.Meta)

		pbLogs = append(pbLogs, pbLog)
	}

	var convertedMeta *proto.Meta

	// Convert meta to proto
	if payload.Meta != nil {
		convertedMeta = protohelper.ConvertMetaToProtoMeta(payload.Meta)
	}

	req := &proto.LogPayload{
		AgentId:    payload.AgentID,
		HostId:     payload.HostID,
		Hostname:   payload.Hostname,
		EndpointId: payload.EndpointID,
		Timestamp:  timestamppb.New(payload.Timestamp),
		Logs:       pbLogs,
		Meta:       convertedMeta,
	}
	// Marshal manually to check for proto errors
	_, err := goproto.Marshal(req)
	if err != nil {
		utils.Error("Failed to marshal proto.LogPayload: %v", err)
		return err
	}
	//utils.Debug("Marshaled LogPayload size = %d bytes", len(b))

	// Try sending
	err = s.stream.Send(req)
	if err != nil {
		utils.Warn("Log stream send failed: %v", err)

		// Check if it's retryable
		if stat, ok := status.FromError(err); ok {
			if stat.Code() == codes.Unavailable || stat.Code() == codes.Canceled || stat.Code() == codes.DeadlineExceeded {
				utils.Warn("Stream error %v detected, reconnecting...", stat.Code())

				if reconnectErr := s.reconnectStream(context.Background()); reconnectErr != nil {
					return reconnectErr
				}

				// After reconnecting, retry sending once
				utils.Debug("Retrying after reconnect...")
				return s.stream.Send(req)
			}
		}

		// Other types of errors: return immediately
		return err
	}

	utils.Debug("Streamed %d logs", len(pbLogs))
	return nil
}

// reconnectStream attempts to reconnect the log stream.
// It closes the existing stream and opens a new one.
// The function logs the status of the reconnection attempt and returns an error if it fails.
// It uses the context passed as a parameter to manage the lifecycle of the stream.
func (s *LogSender) reconnectStream(ctx context.Context) error {
	utils.Warn("Attempting to reconnect log stream...")

	// Close existing stream
	if s.stream != nil {
		_ = s.stream.CloseSend() // best effort
	}

	// Open a new stream
	newStream, err := s.client.SubmitStream(ctx)
	if err != nil {
		utils.Error("Failed to reconnect log stream: %v", err)
		return err
	}

	s.stream = newStream
	utils.Info("Reconnected log stream successfully")
	return nil
}
