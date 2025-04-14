package logrunner

import (
	"context"
	"fmt"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/agent/internal/logs/logcollector"
	"github.com/aaronlmathis/gosight/agent/internal/logs/logsender"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type LogRunner struct {
	Config      *config.Config
	LogSender   *logsender.LogSender
	LogRegistry *logcollector.LogRegistry
}

func NewRunner(ctx context.Context, cfg *config.Config, agentID string) (*LogRunner, error) {
	logRegistry := logcollector.NewRegistry(cfg)
	logSender, err := logsender.NewSender(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create sender: %v", err)
	}

	return &LogRunner{
		Config:      cfg,
		LogSender:   logSender,
		LogRegistry: logRegistry,
	}, nil
}

func (r *LogRunner) Run(ctx context.Context) {
	defer r.LogSender.Close()

	taskQueue := make(chan *model.LogPayload, 500)
	go r.LogSender.StartWorkerPool(ctx, taskQueue, 10)

	ticker := time.NewTicker(r.Config.Agent.Interval)
	defer ticker.Stop()

	utils.Info("Log Runner started. Sending logs every %v", r.Config.Agent.Interval)

	for {
		select {
		case <-ctx.Done():
			utils.Warn("agent shutting down...")
			return
		case <-ticker.C:
			logEntries, err := r.LogRegistry.Collect(ctx)
			if err != nil {
				fmt.Printf("log collection failed: %v", err)
				continue
			}

			// Group logs into payloads by host
			// (or modify to send all logs in one payload if you prefer batching)
			for _, log := range logEntries {
				// Assign endpoint ID
				log.Meta.EndpointID = "temp value" // TODO utils.GenerateEndpointIDLog(log.Meta)

				payload := &model.LogPayload{
					EndpointID: log.Meta.EndpointID,
					Timestamp:  log.Timestamp,
					Logs:       []model.LogEntry{log},
					Meta:       log.Meta,
				}

				select {
				case taskQueue <- payload:
				default:
					utils.Warn("⚠️ Log task queue full! Dropping log from host %s", log.Host)
				}
			}
		}
	}
}
