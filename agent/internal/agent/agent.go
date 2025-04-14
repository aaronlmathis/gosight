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
	agentidentity "github.com/aaronlmathis/gosight/agent/internal/identity"
	metricrunner "github.com/aaronlmathis/gosight/agent/internal/metrics/metricrunner"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type Agent struct {
	Config       *config.Config
	MetricRunner *metricrunner.MetricRunner
	AgentID      string
	AgentVersion string
}

func NewAgent(ctx context.Context, cfg *config.Config, agentVersion string) (*Agent, error) {
	agentID, err := agentidentity.LoadOrCreateAgentID()
	if err != nil {
		utils.Fatal("Failed to get agent ID: %v", err)
	}
	metricRunner, err := metricrunner.NewRunner(ctx, cfg, agentID, agentVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %v", err)
	}
	return &Agent{
		Config:       cfg,
		MetricRunner: metricRunner,
		AgentID:      agentID,
		AgentVersion: agentVersion,
	}, nil
}

func (a *Agent) Start(ctx context.Context) {

	// Start runner.
	a.MetricRunner.Run(ctx)
}

func (a *Agent) Close(ctx context.Context) {
	// Stop the metric runner
	a.MetricRunner.Close()

	utils.Info("Agent shutdown complete")

}
