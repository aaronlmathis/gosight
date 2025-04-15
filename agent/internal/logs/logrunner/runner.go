package logrunner

import (
	"context"
	"fmt"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/agent/internal/logs/logcollector"
	"github.com/aaronlmathis/gosight/agent/internal/logs/logsender"
	"github.com/aaronlmathis/gosight/agent/internal/meta"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type LogRunner struct {
	Config       *config.Config
	LogSender    *logsender.LogSender
	LogRegistry  *logcollector.LogRegistry
	AgentID      string
	AgentVersion string
}

func NewRunner(ctx context.Context, cfg *config.Config, agentID, agentVersion string) (*LogRunner, error) {
	logRegistry := logcollector.NewRegistry(cfg)
	logSender, err := logsender.NewSender(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create sender: %v", err)
	}

	return &LogRunner{
		Config:       cfg,
		LogSender:    logSender,
		LogRegistry:  logRegistry,
		AgentID:      agentID,
		AgentVersion: agentVersion,
	}, nil
}

func (r *LogRunner) Close() {
	if r.LogSender != nil {
		_ = r.LogSender.Close()
	}
}

func (r *LogRunner) Run(ctx context.Context) {

	defer r.LogSender.Close()
	utils.Debug("LogRunner started, but queue not so much")
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
			// Build standard host meta first, to GenerateEndpointID.
			meta := meta.BuildHostMeta(r.Config, nil, r.AgentID, r.AgentVersion)
			meta.AgentID = r.AgentID
			meta.AgentVersion = r.AgentVersion
			// Generate EndpointID
			endpointID := utils.GenerateEndpointID(meta)
			// Set Meta EndpointID Field.
			meta.EndpointID = endpointID

			var logsToPackage []model.LogEntry

			for _, log := range logEntries {

				if log.Meta == nil {
					utils.Warn("LogEntry missing LogMeta: %+v", log)
					log.Meta = &model.LogMeta{AppName: "unknown"}
				}

				log.Meta.AgentID = r.AgentID

				logsToPackage = append(logsToPackage, log)
			}

			// Step 3: Wrap in LogPayload
			payload := &model.LogPayload{
				EndpointID: meta.EndpointID,
				Timestamp:  time.Now(),
				Logs:       logsToPackage,
				Meta:       meta,
			}

			utils.Debug("Log Payload: %+v", payload)

			select {
			case taskQueue <- payload:
			default:
				utils.Warn(" Log task queue full! Dropping log from host %s", meta.Hostname)
			}
		}
	}
}
