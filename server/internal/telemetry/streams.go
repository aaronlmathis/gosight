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

// gosight/server/internal/api
// streams.go - gRPC handler for stream submission

package telemetry

import (
	"time"

	"github.com/aaronlmathis/gosight/server/internal/sys"
	"github.com/aaronlmathis/gosight/shared/model"
	pb "github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
	"google.golang.org/protobuf/proto"
)

// StreamHandler implements pb.MetricsServiceServer
// StreamHandler implements MetricsServiceServer
type StreamHandler struct {
	Sys *sys.SystemContext
	pb.UnimplementedStreamServiceServer
}

func NewStreamHandler(sys *sys.SystemContext) *StreamHandler {
	utils.Debug("StreamHandler initialized with store: %T", sys.Stores.Metrics)
	return &StreamHandler{
		Sys: sys,
	}
}

func (h *StreamHandler) Stream(stream pb.StreamService_StreamServer) error {
	var registered bool
	var agentID string

	go func() {
		for {
			time.Sleep(5 * time.Second)
			utils.Debug("server Stream handler alive...")
		}
	}()

	for {

		req, err := stream.Recv()
		if err != nil {
			utils.Error("Stream receive error: %v", err)
			return err
		}

		switch v := req.Payload.(type) {

		case *pb.StreamPayload_Process:
			SafeHandlePayload(func() {
				var processPayload pb.ProcessPayload
				if err := proto.Unmarshal(v.Process.RawPayload, &processPayload); err != nil {
					utils.Warn("Failed to unmarshal ProcessPayload: %v", err)
				}
				// Convert the protobuf ProcessPayload to model.ProcessPayload
				converted := ConvertProtoProcessPayload(&processPayload)

				// Tag enrichment
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

				h.Sys.Cache.Processes.Add(model.ProcessSnapshot{
					Timestamp:  converted.Timestamp,
					HostID:     converted.HostID,
					EndpointID: converted.EndpointID,
					Processes:  converted.Processes,
					Meta:       converted.Meta,
				})

				// Broadcast + store process payload

				utils.Debug("📡 About to broadcast process payload for %s with %d processes", converted.EndpointID, len(converted.Processes))
				h.Sys.WSHub.Processes.Broadcast(converted)

				/*				Holding off on this for now, as it is simply too much to store this in the database.

								if err := h.Sys.Stores.Processes.Write([]model.ProcessPayload{converted}); err != nil {
								// Insert Process Snapshot and ProcessInfos into database.
								if err := h.Sys.Stores.Data.InsertFullProcessPayload(stream.Context(), &converted); err != nil {
									utils.Warn("Failed to store ProcessPayload: %v", err)
									return
								} else {
									utils.Debug("Stored ProcessPayload from agent %s", converted.EndpointID)
								}
				*/
			})
		case *pb.StreamPayload_Metric:
			// Handle MetricWrapper
			SafeHandlePayload(func() {
				var metricPayload pb.MetricPayload
				if err := proto.Unmarshal(v.Metric.RawPayload, &metricPayload); err != nil {
					utils.Error("Failed to unmarshal MetricPayload: %v", err)
					return
				}

				converted := ConvertToModelPayload(&metricPayload)

				//utils.Debug("Received MetricPayload from agent %s: %s", converted.EndpointID, converted.Meta.AgentID)

				// Register agent session
				if !registered && converted.Meta != nil && converted.Meta.AgentID != "" {
					agentID = converted.AgentID
					h.Sys.Tracker.RegisterAgentSession(agentID, stream)
					//utils.Info("Registered live session for agent %s (%s)", converted.Meta.Hostname, agentID)
					registered = true
				}

				// Tag enrichment
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

				// Evaluate rules
				h.Sys.Tele.Evaluator.EvaluateMetric(h.Sys.Ctx, converted.Metrics, converted.Meta)

				// Update agent/container info
				h.Sys.Tracker.UpdateAgent(converted.Meta)
				if converted.Meta.ContainerID != "" {
					h.Sys.Tracker.UpdateContainer(converted.Meta)
				}

				// Broadcast + store metrics
				h.Sys.WSHub.Metrics.Broadcast(converted)
				if err := h.Sys.Stores.Metrics.Write([]model.MetricPayload{converted}); err != nil {
					utils.Warn("Failed to enqueue metrics from %s: %v", converted.EndpointID, err)
				} else {

					// TODO - Need to merge meta + dimensions before sending to metric store, that way it can be done once and sent to both metric index and metric store.
					// This is a temporary fix.
					if converted.Meta != nil && converted.Meta.EndpointID != "" {
						h.Sys.Tele.Meta.Set(converted.Meta.EndpointID, *converted.Meta)
					}
					for _, m := range converted.Metrics {
						merged := MergeDimensionsWithMeta(m.Dimensions, converted.Meta)
						h.Sys.Tele.Index.Add(m.Namespace, m.SubNamespace, m.Name, merged)

					}
				}
			})

		case *pb.StreamPayload_CommandResponse:
			// Received command execution result
			//utils.Info("Received CommandResponse: success=%v output=%s error=%s",	v.CommandResponse.Success, v.CommandResponse.Output, v.CommandResponse.ErrorMessage)

			if endpointID, ok := h.Sys.Tracker.GetEndpointIdByAgentId(agentID); ok {
				h.Sys.WSHub.Commands.Broadcast(&model.CommandResult{

					EndpointID:   endpointID,
					Output:       v.CommandResponse.Output,
					Success:      v.CommandResponse.Success,
					ErrorMessage: v.CommandResponse.ErrorMessage,
					Timestamp:    time.Now().Format(time.RFC3339),
				})
			} else {
				utils.Warn("No endpoint ID found for agent %s", agentID)
			}

		case *pb.StreamPayload_CommandRequest:
			// Server shouldn't receive CommandRequests — log it
			utils.Warn("Unexpected CommandRequest received from agent")

		default:
			utils.Warn("Unknown payload type received on stream")
		}

		//  Always send a StreamResponse after processing
		resp := &pb.StreamResponse{
			Status:     "ok",
			StatusCode: 0,
		}

		//utils.Debug("Agent ID: %s", agentID)
		var pendingCmd *pb.CommandRequest
		if agentID != "" {
			pendingCmd = h.Sys.Tracker.DequeueCommand(agentID) // Copy command reference outside of lock
		}
		if pendingCmd != nil {
			resp.Command = pendingCmd
			utils.Info("Injecting pending CommandRequest into StreamResponse for agent %s", agentID)
		}
		//utils.Debug("Sending StreamResponse to %s", agentID)
		if session, ok := h.Sys.Tracker.GetAgentSession(agentID); ok {
			select {
			case session.SendQueue <- resp:
				// sent successfully
			default:
				utils.Warn("SendQueue full for agent %s — dropping response", agentID)
			}
		} else {
			//utils.Warn("No live session found for agent %s", agentID)
		}
	}
}
