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
// Package model contains the data structures used in GoSight.

// agent/processes/processsender/sender.go

package processsender

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	grpcconn "github.com/aaronlmathis/gosight/agent/internal/grpc"
	"github.com/aaronlmathis/gosight/agent/internal/protohelper"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
	goproto "google.golang.org/protobuf/proto"

	"google.golang.org/grpc"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProcessSender struct {
	cfg    *config.Config
	ctx    context.Context
	cc     *grpc.ClientConn
	client proto.StreamServiceClient
	stream proto.StreamService_StreamClient
	wg     sync.WaitGroup
	streamCtx    context.Context
	streamCancel context.CancelFunc
	onDisconnect func()
}

func NewSender(ctx context.Context, cfg *config.Config) (*ProcessSender, error) {
	cc, err := grpcconn.GetGRPCConn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	client := proto.NewStreamServiceClient(cc)
	streamCtx, cancel := context.WithCancel(ctx)
	stream, err := client.Stream(streamCtx)
	if err != nil {
		cancel()
		return nil, err
	}

	return &ProcessSender{
		cfg:    cfg,
		ctx:    ctx,
		cc:     cc,
		client: client,
		stream: stream,
		streamCtx:    streamCtx,
		streamCancel: cancel,
	}, nil
}

func (s *ProcessSender) SetDisconnectHandler(fn func()) {
	s.onDisconnect = fn
}

func (s *ProcessSender) Close() error {
	utils.Info("Closing ProcessSender...")

	// Cancel stream context to stop any active Send operations
	if s.streamCancel != nil {
		s.streamCancel()
	}

	// Wait for background workers to finish
	s.wg.Wait()
	utils.Info("All ProcessSender workers finished")

	// Close gRPC connection
	if s.cc != nil {
		if err := s.cc.Close(); err != nil {
			utils.Warn("Error closing gRPC connection: %v", err)
			return err
		}
	}

	utils.Info("ProcessSender closed successfully")
	return nil
}

func (s *ProcessSender) SendSnapshot(payload *model.ProcessPayload) error {
	pb := &proto.ProcessPayload{
		AgentId:    payload.AgentID,
		HostId:     payload.HostID,
		Hostname:   payload.Hostname,
		EndpointId: payload.EndpointID,
		Timestamp:  timestamppb.New(payload.Timestamp),
		Meta:       protohelper.ConvertMetaToProtoMeta(payload.Meta),
		
	}

	for _, p := range payload.Processes {
		pb.Processes = append(pb.Processes, &proto.ProcessInfo{
			Pid:        int32(p.PID),
			Ppid:       int32(p.PPID),
			User:       p.User,
			Executable: p.Executable,
			Cmdline:    p.Cmdline,
			CpuPercent: p.CPUPercent,
			MemPercent: p.MemPercent,
			Threads:    int32(p.Threads),
			StartTime:  timestamppb.New(p.StartTime),
			Tags:       p.Tags,
		})
	}

	b, err := goproto.Marshal(pb)
	if err != nil {
		return fmt.Errorf("marshal ProcessPayload: %w", err)
	}

	sp := &proto.StreamPayload{
		Payload: &proto.StreamPayload_Process{
			Process: &proto.ProcessWrapper{
				RawPayload: b,
			},
		},
	}

	const maxAttempts = 5
	var backoff = []time.Duration{500 * time.Millisecond, 1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if s.streamCtx.Err() != nil {
			utils.Warn("ProcessSender context canceled, aborting SendSnapshot")
			if s.onDisconnect != nil {
				go s.onDisconnect()
			}
			return fmt.Errorf("stream context canceled")
		}
		sendCtx, cancel := context.WithTimeout(s.streamCtx, 5*time.Second)
		sendCh := make(chan error, 1)
		go func() {
			sendCh <- s.stream.Send(sp)
		}()

		select {
		case err := <-sendCh:
			cancel()
			if err != nil {
				utils.Warn("Unknown process send error — retrying in %v [attempt %d/%d]: %v", backoff[attempt-1], attempt, maxAttempts, err)
				if recErr := s.reconnectStream(); recErr != nil {
					utils.Error("Failed to reconnect process stream: %v", recErr)
					return fmt.Errorf("send failed and reconnect failed: %w", err)
				}
				time.Sleep(backoff[attempt-1])
				continue
			}
			return nil
		case <-sendCtx.Done():
			cancel()
			utils.Warn("Process send timed out — retrying in %v [attempt %d/%d]", backoff[attempt-1], attempt, maxAttempts)
			if recErr := s.reconnectStream(); recErr != nil {
				utils.Error("Failed to reconnect process stream: %v", recErr)
				return fmt.Errorf("send timeout and reconnect failed")
			}
			time.Sleep(backoff[attempt-1])
		}
	}

	utils.Error("All process send attempts failed, triggering onDisconnect")
	if s.onDisconnect != nil {
		go s.onDisconnect()
	}

	return fmt.Errorf("send failed after %d attempts: EOF", maxAttempts)
}


func (s *ProcessSender) reconnectStream() error {
	utils.Warn("Attempting to reconnect process stream...")

	// Close old connection and cancel old stream context
	if s.streamCancel != nil {
		s.streamCancel()
	}
	if s.cc != nil {
		_ = s.cc.Close()
	}

	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	cc, err := grpcconn.GetGRPCConn(ctx, s.cfg)
	if err != nil {
		return err
	}

	client := proto.NewStreamServiceClient(cc)
	streamCtx, streamCancel := context.WithCancel(s.ctx)
	stream, err := client.Stream(streamCtx)
	if err != nil {
		streamCancel() // avoid leaking context
		return err
	}

	// Replace old references
	s.cc = cc
	s.client = client
	s.stream = stream
	s.streamCtx = streamCtx
	s.streamCancel = streamCancel

	utils.Info("Reconnected process stream successfully")
	return nil
}