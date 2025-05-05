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

// internal/agent.go
// gosight/agent/internal/agent.go

package gosightagent

import (
	"context"
	"fmt"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	grpcconn "github.com/aaronlmathis/gosight/agent/internal/grpc"
	agentidentity "github.com/aaronlmathis/gosight/agent/internal/identity"
	"github.com/aaronlmathis/gosight/agent/internal/logs/logrunner"
	"github.com/aaronlmathis/gosight/agent/internal/meta"
	metricrunner "github.com/aaronlmathis/gosight/agent/internal/metrics/metricrunner"
	"github.com/aaronlmathis/gosight/agent/internal/processes/processrunner"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
	"google.golang.org/grpc"
)

type Agent struct {
	Config        *config.Config
	MetricRunner  *metricrunner.MetricRunner
	AgentID       string
	AgentVersion  string
	LogRunner     *logrunner.LogRunner
	ProcessRunner *processrunner.ProcessRunner
	Meta          *model.Meta
	Ctx           context.Context
}

func NewAgent(ctx context.Context, cfg *config.Config, agentVersion string) (*Agent, error) {
	agentID, err := agentidentity.LoadOrCreateAgentID()
	if err != nil {
		utils.Fatal("Failed to get agent ID: %v", err)
	}

	baseMeta := meta.BuildMeta(cfg, nil, agentID, agentVersion)

	backoff := 1 * time.Second
	maxBackoff := 15 * time.Minute

	var conn *grpc.ClientConn

	for {
		utils.Info("Attempting to connect to server at %s...", cfg.Agent.ServerURL)

		conn, err = grpcconn.GetGRPCConn(ctx, cfg)
		if err != nil {
			utils.Warn("gRPC dial failed: %v", err)
		} else {
			// Validate stream can be established
			client := proto.NewStreamServiceClient(conn)
			streamCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			_, streamErr := client.Stream(streamCtx)
			cancel()

			if streamErr == nil {
				utils.Info("Successfully established gRPC connection and stream.")
				break
			}

			utils.Warn("gRPC dial succeeded but stream failed: %v", streamErr)
		}

		utils.Warn("Retrying connection in %v...", backoff)
		select {
		case <-time.After(backoff):
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Now that connection is confirmed usable, initialize runners
	metricRunner, err := metricrunner.NewRunner(ctx, cfg, baseMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric runner: %v", err)
	}

	logRunner, err := logrunner.NewRunner(ctx, cfg, baseMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to create log runner: %v", err)
	}

	processRunner, err := processrunner.NewRunner(ctx, cfg, baseMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to create process runner: %v", err)
	}

	return &Agent{
		Ctx:          ctx,
		Config:       cfg,
		AgentID:      agentID,
		AgentVersion: agentVersion,
		MetricRunner: metricRunner,
		LogRunner:    logRunner,
		ProcessRunner: processRunner,
		Meta:         baseMeta,
	}, nil
}

func (a *Agent) Start(ctx context.Context) {

	// Start runner.
	utils.Debug("Agent attempting to start metricrunner.")
	go a.MetricRunner.Run(ctx)

	utils.Debug("Agent attempting to start metricrunner.")
	go a.LogRunner.Run(ctx)

	utils.Debug("Agent attempting to start processrunner.")
	go a.ProcessRunner.Run(ctx)

}

func (a *Agent) Close(ctx context.Context) {
	// Stop All Runners
	a.MetricRunner.Close()
	a.LogRunner.Close()
	a.ProcessRunner.Close()

	err := grpcconn.CloseGRPCConn()
	if err != nil {
		utils.Warn("Failed to close gRPC connection cleanly: %v", err)
	}
	utils.Info("Agent shutdown complete")

}
