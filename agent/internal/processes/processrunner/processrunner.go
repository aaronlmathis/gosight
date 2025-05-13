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

// ProcessRunner is a struct that handles the collection and sending of process data.
// It manages the process collection interval, the task queue, and the
// process sender. It implements the Run method to start the collection process
// and the Close method to clean up resources.
type ProcessRunner struct {
	Config        *config.Config
	ProcessSender *processsender.ProcessSender
	Meta          *model.Meta
}

// NewRunner creates a new ProcessRunner instance.
// It initializes the process sender and sets up the context for the runner.
// It returns a pointer to the ProcessRunner and an error if any occurs during initialization.
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

// SetDisconnectHandler sets a handler function to be called when the process sender disconnects.
// This allows the user to define custom behavior when the connection to the server is lost.
// The handler function will be called when the disconnect event occurs.
func (r *ProcessRunner) SetDisconnectHandler(fn func()) {
	r.ProcessSender.SetDisconnectHandler(fn)
}

// Close closes the process sender and cleans up resources.
// It ensures that all resources are released and the connection to the server is properly closed.
// This method should be called when the process runner is no longer needed.
func (r *ProcessRunner) Close() {
	if r.ProcessSender != nil {
		_ = r.ProcessSender.Close()
	}
}

// Run starts the process collection and sending loop.
// It collects process data at the specified interval and sends it to the server.
// The method runs indefinitely until the context is done or an error occurs.
// It uses a ticker to trigger the collection process at regular intervals.
// The collected process data is sent to a task queue, which is processed by a worker pool.
// The worker pool is managed by the ProcessSender, which handles the sending of data to the server.
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
