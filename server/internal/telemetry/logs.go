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
	for {
		pbPayload, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.LogResponse{
				Status:     "ok",
				StatusCode: 0,
			})
		}
		if err != nil {
			utils.Error("Log stream receive error: %v", err)
			return err
		}

		// Convert payload into a model.LogPayload.
		payload := ConvertToModelLogPayload(pbPayload)

		// Websocket broadcast
		h.Sys.Web.BroadcastLog(payload)
		err = h.Sys.Stores.Logs.Write([]model.LogPayload{payload}, stream.Context())
		if err != nil {
			utils.Error("Failed to store logs from host %s: %v", payload.EndpointID, err)
		} else {
			utils.Info("Stored %d logs from host: %s at %s", len(payload.Logs), payload.EndpointID, payload.Timestamp)
		}
	}
}
