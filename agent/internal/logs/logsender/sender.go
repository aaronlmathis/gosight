package logsender

import (
	"context"
	"fmt"
	"sync"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	grpcconn "github.com/aaronlmathis/gosight/agent/internal/grpc"
	"github.com/aaronlmathis/gosight/agent/internal/protohelper"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	goproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Sender holds the gRPC client and connection
type LogSender struct {
	client proto.LogServiceClient
	cc     *grpc.ClientConn
	stream proto.LogService_SubmitStreamClient
	wg     sync.WaitGroup
	Config *config.Config
}

// NewSender establishes a gRPC connection
func NewSender(ctx context.Context, cfg *config.Config) (*LogSender, error) {
	clientConn, err := grpcconn.GetGRPCConn(ctx, cfg)
	if err != nil {
		return nil, err
	}

	client := proto.NewLogServiceClient(clientConn)
	utils.Info("established gRPC Connection with %v", cfg.Agent.ServerURL)

	//
	stream, err := client.SubmitStream(ctx)
	if err != nil {
		utils.Debug("Failed to open stream: %v", err)
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	return &LogSender{
		client: client,
		cc:     clientConn,
		stream: stream,
		Config: cfg,
	}, nil

}

func (s *LogSender) Close() error {
	utils.Info("Closing LogSender... waiting for workers")
	s.wg.Wait()
	utils.Info("All LogSender workers finished")
	return s.cc.Close()
}

func (s *LogSender) SendLogs(payload *model.LogPayload) error {
	pbLogs := make([]*proto.LogEntry, 0, len(payload.Logs))

	for _, log := range payload.Logs {
		pbLog := &proto.LogEntry{
			Timestamp: timestamppb.New(log.Timestamp),
			Level:     log.Level,
			Message:   log.Message,
			Source:    log.Source,
			Category:  log.Category,
			Pid:       int32(log.PID),
			Fields:    log.Fields,
			Tags:      log.Tags,
			Meta:      protohelper.ConvertLogMetaToProto(log.Meta),
		}
		utils.Debug("Sender: LogEntry: %v", pbLog)
		utils.Debug("Sender:LogMeta: %v", pbLog.Meta)

		pbLogs = append(pbLogs, pbLog)
	}

	var convertedMeta *proto.Meta

	// Convert meta to proto
	if payload.Meta != nil {
		convertedMeta = protohelper.ConvertMetaToProtoMeta(payload.Meta)
	}

	req := &proto.LogPayload{
		AgentId:    payload.AgentID,
		HostId:     payload.HostID,
		Hostname:   payload.Hostname,
		EndpointId: payload.EndpointID,
		Timestamp:  timestamppb.New(payload.Timestamp),
		Logs:       pbLogs,
		Meta:       convertedMeta,
	}
	// Marshal manually to check for proto errors
	b, err := goproto.Marshal(req)
	if err != nil {
		utils.Error("Failed to marshal proto.LogPayload: %v", err)
		return err
	}
	utils.Debug("Marshaled LogPayload size = %d bytes", len(b))

	if err := s.stream.Send(req); err != nil {
		return fmt.Errorf("log stream send failed: %w", err)
	}

	utils.Debug("Streamed %d logs", len(pbLogs))
	// Try sending
	err = s.stream.Send(req)
	if err != nil {
		utils.Warn("Log stream send failed: %v", err)

		// Check if it's retryable
		if stat, ok := status.FromError(err); ok {
			if stat.Code() == codes.Unavailable || stat.Code() == codes.Canceled || stat.Code() == codes.DeadlineExceeded {
				utils.Warn("Stream error %v detected, reconnecting...", stat.Code())

				if reconnectErr := s.reconnectStream(context.Background()); reconnectErr != nil {
					return reconnectErr
				}

				// After reconnecting, retry sending once
				utils.Debug("Retrying after reconnect...")
				return s.stream.Send(req)
			}
		}

		// Other types of errors: return immediately
		return err
	}

	utils.Debug("Streamed %d logs", len(pbLogs))
	return nil
}
func (s *LogSender) reconnectStream(ctx context.Context) error {
	utils.Warn("Attempting to reconnect log stream...")

	// Close existing stream
	if s.stream != nil {
		s.stream.CloseSend() // best effort
	}

	// Open a new stream
	newStream, err := s.client.SubmitStream(ctx)
	if err != nil {
		utils.Error("Failed to reconnect log stream: %v", err)
		return err
	}

	s.stream = newStream
	utils.Info("Reconnected log stream successfully")
	return nil
}
