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

// Package agent provides the main functionality for the GoSight agent.
// It handles the initialization and management of various components
// such as metrics, logs, and processes. The agent is responsible for
// collecting data from the system and sending it to the GoSight server.
// It also manages the agent's identity and configuration.
package gosightagent

import (
	"context"
	"reflect"
	"testing"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/agent/internal/logs/logrunner"
	metricrunner "github.com/aaronlmathis/gosight/agent/internal/metrics/metricrunner"
	"github.com/aaronlmathis/gosight/agent/internal/processes/processrunner"
	"github.com/aaronlmathis/gosight/shared/model"
)

func TestNewAgent(t *testing.T) {
	type args struct {
		ctx          context.Context
		cfg          *config.Config
		agentVersion string
	}
	tests := []struct {
		name    string
		args    args
		want    *Agent
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAgent(tt.args.ctx, tt.args.cfg, tt.args.agentVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAgent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_Start(t *testing.T) {
	type fields struct {
		Config        *config.Config
		MetricRunner  *metricrunner.MetricRunner
		AgentID       string
		AgentVersion  string
		LogRunner     *logrunner.LogRunner
		ProcessRunner *processrunner.ProcessRunner
		Meta          *model.Meta
		Ctx           context.Context
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				Config:        tt.fields.Config,
				MetricRunner:  tt.fields.MetricRunner,
				AgentID:       tt.fields.AgentID,
				AgentVersion:  tt.fields.AgentVersion,
				LogRunner:     tt.fields.LogRunner,
				ProcessRunner: tt.fields.ProcessRunner,
				Meta:          tt.fields.Meta,
				Ctx:           tt.fields.Ctx,
			}
			a.Start(tt.args.ctx)
		})
	}
}

func TestAgent_Close(t *testing.T) {
	type fields struct {
		Config        *config.Config
		MetricRunner  *metricrunner.MetricRunner
		AgentID       string
		AgentVersion  string
		LogRunner     *logrunner.LogRunner
		ProcessRunner *processrunner.ProcessRunner
		Meta          *model.Meta
		Ctx           context.Context
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				Config:        tt.fields.Config,
				MetricRunner:  tt.fields.MetricRunner,
				AgentID:       tt.fields.AgentID,
				AgentVersion:  tt.fields.AgentVersion,
				LogRunner:     tt.fields.LogRunner,
				ProcessRunner: tt.fields.ProcessRunner,
				Meta:          tt.fields.Meta,
				Ctx:           tt.fields.Ctx,
			}
			a.Close()
		})
	}
}
