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
}

func NewSender(ctx context.Context, cfg *config.Config) (*ProcessSender, error) {
	cc, err := grpcconn.GetGRPCConn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	client := proto.NewStreamServiceClient(cc)
	stream, err := client.Stream(ctx)
	if err != nil {
		return nil, err
	}

	return &ProcessSender{
		cfg:    cfg,
		ctx:    ctx,
		cc:     cc,
		client: client,
		stream: stream,
	}, nil
}

func (s *ProcessSender) Close() error {
	s.wg.Wait()
	return s.cc.Close()
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
	utils.Debug("Sending ProcessPayload with %d processes", len(pb.Processes))
	sp := &proto.StreamPayload{
		Payload: &proto.StreamPayload_Process{
			Process: &proto.ProcessWrapper{
				RawPayload: b,
			},
		},
	}

	sendCtx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	sendCh := make(chan error, 1)
	go func() {
		sendCh <- s.stream.Send(sp)
	}()

	select {
	case err := <-sendCh:
		return err
	case <-sendCtx.Done():
		return fmt.Errorf("send timeout")
	}
}
