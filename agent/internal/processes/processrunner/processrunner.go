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
// agent/processes/processrunner/runner.go

package processrunner

import (
	"context"
	"fmt"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/agent/internal/meta"
	"github.com/aaronlmathis/gosight/agent/internal/processes/processcollector"
	"github.com/aaronlmathis/gosight/agent/internal/processes/processsender"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type ProcessRunner struct {
	Config        *config.Config
	ProcessSender *processsender.ProcessSender
	Meta          *model.Meta
}

func NewRunner(ctx context.Context, cfg *config.Config, baseMeta *model.Meta) (*ProcessRunner, error) {
	sender, err := processsender.NewSender(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create process sender: %w", err)
	}
	return &ProcessRunner{
		Config:        cfg,
		ProcessSender: sender,
		Meta:          baseMeta,
	
	}, nil
}

func (r *ProcessRunner) SetDisconnectHandler(fn func()) {
	r.ProcessSender.SetDisconnectHandler(fn)
}

func (r *ProcessRunner) Close() {
	if r.ProcessSender != nil {
		_ = r.ProcessSender.Close()
	}
}

func (r *ProcessRunner) Run(ctx context.Context) {
	taskQueue := make(chan *model.ProcessPayload, 100)
	go r.ProcessSender.StartWorkerPool(ctx, taskQueue, r.Config.Agent.ProcessCollection.Workers)

	ticker := time.NewTicker(r.Config.Agent.ProcessCollection.Interval)
	defer ticker.Stop()

	utils.Info("ProcessRunner started. Collecting processes every %v", r.Config.Agent.ProcessCollection.Interval)

	for {
		select {
		case <-ctx.Done():
			utils.Warn("ProcessRunner shutting down")
			return
		case <-ticker.C:
			snapshot, err := processcollector.CollectProcesses(ctx)
			if err != nil {
				utils.Error("Failed to collect processes: %v", err)
				continue
			}

			metaCopy := meta.CloneMetaWithTags(r.Meta, nil)
			metaCopy.EndpointID = utils.GenerateEndpointID(metaCopy)

			payload := &model.ProcessPayload{
				AgentID:    metaCopy.AgentID,
				HostID:     metaCopy.HostID,
				Hostname:   metaCopy.Hostname,
				EndpointID: metaCopy.EndpointID,
				Timestamp:  snapshot.Timestamp,
				Processes:  snapshot.Processes,
				Meta:       metaCopy,
			}

			select {
			case taskQueue <- payload:
				// ok
			default:
				utils.Warn("Process task queue full. Dropping snapshot")
			}
		}
	}
}
