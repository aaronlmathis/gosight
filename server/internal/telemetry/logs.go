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

		payload := ConvertToModelLogPayload(pbPayload)

		h.Sys.Web.BroadcastLog(payload)
		if err := h.Sys.Stores.Logs.Write([]model.LogPayload{payload}, stream.Context()); err != nil {
			utils.Error("Failed to store logs from host %s: %v", payload.EndpointID, err)
		} else {
			utils.Debug("Stored %d logs from host: %s at %s", len(payload.Logs), payload.EndpointID, payload.Timestamp)
		}
	}
}
