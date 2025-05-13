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
package logrunner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/agent/internal/logs/logcollector"
	"github.com/aaronlmathis/gosight/agent/internal/logs/logsender"
	"github.com/aaronlmathis/gosight/agent/internal/meta"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// LogRunner is a struct that handles the collection and sending of log data.
// It manages the log collection interval, the task queue, and the
// log sender. It implements the Run method to start the collection process
// and the Close method to clean up resources.
type LogRunner struct {
	Config      *config.Config
	LogSender   *logsender.LogSender
	LogRegistry *logcollector.LogRegistry
	Meta        *model.Meta
	runWg       sync.WaitGroup
}

// NewRunner creates a new LogRunner instance.
// It initializes the log sender and sets up the context for the runner.
// It returns a pointer to the LogRunner and an error if any occurs during initialization.
func NewRunner(ctx context.Context, cfg *config.Config, baseMeta *model.Meta) (*LogRunner, error) {

	logRegistry := logcollector.NewRegistry(cfg)

	logSender, err := logsender.NewSender(ctx, cfg)
	if err != nil {
		// Clean up registry if sender fails?
		logRegistry.Close() // Add a Close method to LogRegistry
		return nil, fmt.Errorf("failed to create sender: %v", err)
	}

	return &LogRunner{
		Config:      cfg,
		LogSender:   logSender,
		LogRegistry: logRegistry,
		Meta:        baseMeta,
	}, nil
}

// Close cleans up the resources used by the LogRunner.
// It closes the log sender and the log registry.
// It should be called when the LogRunner is no longer needed.
func (r *LogRunner) Close() {
	utils.Info("Closing Log Runner...")

	// Close collectors first to stop feeding new logs
	if r.LogRegistry != nil {
		r.LogRegistry.Close() // Ensure LogRegistry has a Close method that calls Close on all collectors
	}

	// Close the sender (which should handle its worker pool shutdown)
	if r.LogSender != nil {
		if err := r.LogSender.Close(); err != nil {
			utils.Error("Error closing log sender: %v", err)
		}
	}

	// Wait for sender pool goroutines to finish (if Close doesn't block)
	// Or manage worker shutdown signalling more explicitly if needed.
	// The LogSender's Close method should ideally handle this wait.
	utils.Info("Log Runner closed.")

}

func (r *LogRunner) Run(ctx context.Context) {

	defer r.Close() // Ensure cleanup on exit

	utils.Debug("Initializing LogRunner...")
	taskQueue := make(chan *model.LogPayload, r.Config.Agent.LogCollection.BufferSize)

	// Start sender worker pool
	// Make sure StartWorkerPool handles context cancellation gracefully
	r.runWg.Add(1)
	go func() {
		defer r.runWg.Done()
		r.LogSender.StartWorkerPool(ctx, taskQueue, r.Config.Agent.LogCollection.Workers)
		utils.Debug("Log sender worker pool stopped.")
	}()

	ticker := time.NewTicker(r.Config.Agent.LogCollection.Interval)
	defer ticker.Stop()

	utils.Info("Log Runner started. Collecting logs every %v", r.Config.Agent.LogCollection.Interval)

	// No need for the startTime throttling anymore unless specifically desired
	// startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			utils.Warn("Log runner context cancelled, shutting down...")
			return // Exit Run, defer Close() will be called
		case <-ticker.C:
			// Collect logs from *all* registered collectors via the registry
			logBatches, err := r.LogRegistry.Collect(ctx) // Assuming Collect iterates through collectors
			if err != nil {
				// Log collection errors, but continue running
				utils.Error("Log collection failed: %v", err)
				continue
			}

			// If no logs collected, continue to next tick
			if len(logBatches) == 0 {
				continue
			}

			// set job tag for victoriametrics.
			r.Meta.Tags["job"] = "gosight-logs"

			// clone base meta before modifying it
			meta := meta.CloneMetaWithTags(r.Meta, nil)

			// Generate Endpoint ID
			endpointID := utils.GenerateEndpointID(meta)
			meta.EndpointID = endpointID
			meta.Tags["instance"] = meta.Hostname

			//utils.Debug("Processing %d log batches for sending.", len(logBatches))

			// Loop through batches collected (potentially from multiple sources)
			for _, batch := range logBatches {
				if len(batch) == 0 {
					continue // Skip empty batches
				}

				// Attach metadata (LogRunner is responsible for the payload structure)
				payload := &model.LogPayload{
					AgentID:    meta.AgentID,
					HostID:     meta.HostID,
					Hostname:   meta.Hostname,
					EndpointID: meta.EndpointID,
					Timestamp:  time.Now(), // Payload timestamp is collection time
					Logs:       batch,      // The batch collected from a specific source
					Meta:       meta,       // Agent/Host metadata
				}

				// No need for the artificial sleep throttling unless rate limiting is required
				// if time.Since(startTime) < 30*time.Second {
				//     time.Sleep(100 * time.Millisecond)
				// }
				//utils.Debug("Queuing log payload with %d entries from host %s", len(batch), meta.Hostname)

				// Send payload to the worker pool queue
				select {
				case taskQueue <- payload:
					// Successfully queued
				case <-ctx.Done():
					utils.Warn("Context cancelled while trying to queue log payload. Shutting down.")
					return // Exit if context cancelled during queuing attempt
				default:
					// Queue is full, drop the batch
					utils.Warn("Log task queue full! Dropping log batch (%d entries) from host %s", len(batch), meta.Hostname)
				}
			}
		}
	}
}
