package telemetry

import (
	"io"

	"github.com/aaronlmathis/gosight/server/internal/sys"
	"github.com/aaronlmathis/gosight/shared/model"
	pb "github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
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

			// Check rulesrunser
			h.Sys.Tele.Evaluator.EvaluateLogs(h.Sys.Ctx, converted.Logs, converted.Meta)

			h.Sys.WSHub.Logs.Broadcast(converted)
			if err := h.Sys.Stores.Logs.Write([]model.LogPayload{converted}, stream.Context()); err != nil {
				utils.Error("Failed to store logs from host %s: %v", converted.EndpointID, err)
			} else {
				utils.Debug("Stored %d logs from host: %s at %s", len(converted.Logs), converted.EndpointID, converted.Timestamp)
			}
		})
	}
}
