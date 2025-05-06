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

package telemetry

import (
	"io"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/events"
	"github.com/aaronlmathis/gosight/server/internal/sys"
	"github.com/aaronlmathis/gosight/shared/model"
	pb "github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/google/uuid"
)

type LogsHandler struct {
	Sys *sys.SystemContext
	pb.UnimplementedLogServiceServer
}

func NewLogsHandler(sys *sys.SystemContext) *LogsHandler {
	utils.Debug("LogsHandler initialized with store: %T", sys.Stores.Logs)
	return &LogsHandler{
		Sys: sys,
	}
}

func (h *LogsHandler) SubmitStream(stream pb.LogService_SubmitStreamServer) error {
	utils.Info("Log SubmitStream started...")

	for {
		pbPayload, err := stream.Recv()
		if err == io.EOF {
			utils.Info("Log stream closed cleanly by client.")
			return nil
		}
		if err != nil {
			utils.Error("Log stream receive error: %v", err)
			return err
		}
		SafeHandlePayload(func() {
			converted := ConvertToModelLogPayload(pbPayload)

			// Enrich Meta.Tags with custom tags from datastore
			if converted.Meta != nil && converted.Meta.EndpointID != "" {
				tags, err := h.Sys.Stores.Data.GetTags(stream.Context(), converted.Meta.EndpointID)
				if err == nil {
					if converted.Meta.Tags == nil {
						converted.Meta.Tags = make(map[string]string)
					}
					for k, v := range tags {
						if _, exists := converted.Meta.Tags[k]; !exists {
							converted.Meta.Tags[k] = v
						}
					}
				}
			}

			// Evaluate severity level of logs and act accordingly
			h.EvaluateSeverityLevel(&converted)

			// Check rulesrunner
			h.Sys.Tele.Evaluator.EvaluateLogs(h.Sys.Ctx, converted.Logs, converted.Meta)

			// Broadcast to hub.LogHub Websocket
			h.Sys.WSHub.Logs.Broadcast(converted)

			// Write to BufferedLog store, fallback to writing to logstore directly.
			if h.Sys.Buffers == nil || h.Sys.Buffers.Metrics == nil {
				utils.Warn("[stream] Logs buffer not configured — writing directly to store")
				// Fall back to writing directly to store
				if err := h.Sys.Stores.Logs.Write([]model.LogPayload{converted}); err != nil {
					utils.Warn("Failed to write logs directly to store: %v", err)
				}
			} else {
				if err := h.Sys.Buffers.Logs.WriteAny(converted); err != nil {
					utils.Warn("Failed to buffer LogPayload: %v", err)
				}	
			}

		})
	}
}

// EvaluateSeverityLevel evaluates the severity level of logs based on thresholds defined in the system.
// Based on that severity, different actions can be taken such as generating events that can trigger alerts.
func (h *LogsHandler) EvaluateSeverityLevel(logPayload *model.LogPayload) {
	// This function can be used to evaluate the severity level of logs
	// based on thresholds defined in the system.

	// Iterate through Log Entries in logPayload
	for _, logEntry := range logPayload.Logs {
		// Example logic to evaluate severity level
		// This is a placeholder and should be replaced with actual logic
		switch logEntry.Level {
		case "critical", "error":
			// Trigger alert or event for error logs
			utils.Info("Error log detected: %s", logEntry.Message)
			// h.Sys.EventManager.TriggerAlert(logEntry) // Example function call
			evt := model.EventEntry{
				ID:         uuid.NewString(),
				Timestamp:  time.Now(),
				Type:       "log.critical",
				Level:      "critical",
				Category:   "logs",
				Message:    utils.Truncate(logEntry.Message, 256),
				Source:     "logs." + logEntry.Source,
				Scope:      "endpoint",
				Target:     logPayload.EndpointID,
				EndpointID: logPayload.EndpointID,
				Meta: events.BuildLogEventMeta(&logEntry, logPayload),
			}
			h.Sys.Tele.Emitter.Emit(h.Sys.Ctx, evt)
		case "warning":
			// Handle warning logs if needed
			//utils.Info("Warning log detected: %s", logEntry.Message)
		case "info":
			// Info logs can be ignored or logged as needed
			//utils.Debug("Info log: %s", logEntry.Message)
		default:
			utils.Debug("Log with unknown level: %s", logEntry.Message)
		}
	}
}