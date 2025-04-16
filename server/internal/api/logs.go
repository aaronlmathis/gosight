package api

// TODO - this is weird package name for these handlers.

import (
	"io"

	"github.com/aaronlmathis/gosight/server/internal/http/websocket"
	"github.com/aaronlmathis/gosight/server/internal/store/logstore"
	"github.com/aaronlmathis/gosight/shared/model"
	pb "github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type LogsHandler struct {
	logstore logstore.LogStore
	pb.UnimplementedLogServiceServer
	websocket *websocket.Hub
}

func NewLogsHandler(s logstore.LogStore, ws *websocket.Hub) *LogsHandler {
	utils.Debug("LogsHandler initialized with store: %T", s)
	return &LogsHandler{
		logstore:  s,
		websocket: ws,
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

		payload := ConvertToModelLogPayload(pbPayload)

		// Websocket broadcast
		h.websocket.Broadcast(websocket.BroadcastEnvelope{
			Type: "logs",
			Data: payload,
		})
		err = h.logstore.Write([]model.LogPayload{payload}, stream.Context())
		if err != nil {
			utils.Error("Failed to store logs from host %s: %v", payload.EndpointID, err)
		} else {
			utils.Info("Stored %d logs from host: %s at %s", len(payload.Logs), payload.EndpointID, payload.Timestamp)
		}
	}
}
