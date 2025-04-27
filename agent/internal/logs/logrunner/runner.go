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

type LogRunner struct {
	Config      *config.Config
	LogSender   *logsender.LogSender
	LogRegistry *logcollector.LogRegistry
	Meta        *model.Meta
	runWg       sync.WaitGroup
}

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

	ticker := time.NewTicker(r.Config.Agent.Interval)
	defer ticker.Stop()

	utils.Info("Log Runner started. Collecting logs every %v", r.Config.Agent.Interval)

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

			// clone base meta before modifying it
			meta := meta.CloneMetaWithTags(r.Meta, nil)

			// Generate Endpoint ID
			endpointID := utils.GenerateEndpointID(meta)
			meta.EndpointID = endpointID

			utils.Debug("Processing %d log batches for sending.", len(logBatches))

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
				utils.Debug("Queuing log payload with %d entries from host %s", len(batch), meta.Hostname)

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
