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

	"github.com/aaronlmathis/gosight/agent/internal/config"
	grpcconn "github.com/aaronlmathis/gosight/agent/internal/grpc"
	agentidentity "github.com/aaronlmathis/gosight/agent/internal/identity"
	"github.com/aaronlmathis/gosight/agent/internal/logs/logrunner"
	"github.com/aaronlmathis/gosight/agent/internal/meta"
	metricrunner "github.com/aaronlmathis/gosight/agent/internal/metrics/metricrunner"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type Agent struct {
	Config       *config.Config
	MetricRunner *metricrunner.MetricRunner
	AgentID      string
	AgentVersion string
	LogRunner    *logrunner.LogRunner
	Meta         *model.Meta
	Ctx          context.Context
}

func NewAgent(ctx context.Context, cfg *config.Config, agentVersion string) (*Agent, error) {

	// Retrieve (or set) the agent ID
	agentID, err := agentidentity.LoadOrCreateAgentID()
	if err != nil {
		utils.Fatal("Failed to get agent ID: %v", err)
	}

	// Build base metadata for the agent and cache it in the Agent struct
	baseMeta := meta.BuildMeta(cfg, nil, agentID, agentVersion)

	metricRunner, err := metricrunner.NewRunner(ctx, cfg, baseMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric runner: %v", err)
	}
	logRunner, err := logrunner.NewRunner(ctx, cfg, baseMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to create log runner: %v", err)
	}
	return &Agent{
		Ctx:          ctx,
		Config:       cfg,
		MetricRunner: metricRunner,
		AgentID:      agentID,
		AgentVersion: agentVersion,
		LogRunner:    logRunner,
		Meta:         baseMeta,
	}, nil
}

func (a *Agent) Start(ctx context.Context) {

	// Start runner.
	utils.Debug("Agent attempting to start metricrunner.")
	go a.MetricRunner.Run(ctx)

	utils.Debug("Agent attempting to start metricrunner.")
	go a.LogRunner.Run(ctx)
}

func (a *Agent) Close(ctx context.Context) {
	// Stop the metric runner
	a.MetricRunner.Close()
	a.LogRunner.Close()

	err := grpcconn.CloseGRPCConn()
	if err != nil {
		utils.Warn("Failed to close gRPC connection cleanly: %v", err)
	}
	utils.Info("Agent shutdown complete")

}
