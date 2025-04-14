package logsender

import (
	"context"
	"fmt"
	"sync"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Sender holds the gRPC client and connection
type LogSender struct {
	client proto.LogServiceClient
	cc     *grpc.ClientConn
	stream proto.LogService_SubmitStreamClient
	wg     sync.WaitGroup
}

// NewSender establishes a gRPC connection
func NewSender(ctx context.Context, cfg *config.Config) (*LogSender, error) {
	// Load TLS config for agent
	tlsCfg, err := agentutils.LoadTLSConfig(cfg)
	if err != nil {
		return nil, err
	}

	// add mTLS to degug log.
	if len(tlsCfg.Certificates) > 0 {
		utils.Info("using mTLS for agent authentication")
	} else {
		utils.Info("Using TLS only (no client certificate)")
	}

	// Establish gRPC connection
	clientConn, err := grpc.NewClient(cfg.Agent.ServerURL,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)),
	)
	if err != nil {
		return nil, err
	}
	utils.Info("connecting to server at: %s", cfg.Agent.ServerURL)
	// Create gRPC client
	// and establish a stream for sending metrics
	client := proto.NewLogServiceClient(clientConn)
	utils.Info("established gRPC Connection with %v", cfg.Agent.ServerURL)

	//
	stream, err := client.SubmitStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	return &LogSender{
		client: client,
		cc:     clientConn,
		stream: stream,
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
			Host:      log.Host,
			Pid:       int32(log.PID),
			Fields:    log.Fields,
			Tags:      log.Tags,
		}

		// If meta is present, convert it
		if log.Meta != nil {
			pbLog.Meta = &proto.LogMeta{
				Os:            log.Meta.OS,
				Platform:      log.Meta.Platform,
				AppName:       log.Meta.AppName,
				AppVersion:    log.Meta.AppVersion,
				ContainerId:   log.Meta.ContainerID,
				ContainerName: log.Meta.ContainerName,
				Unit:          log.Meta.Unit,
				Service:       log.Meta.Service,
				EventId:       log.Meta.EventID,
				User:          log.Meta.User,
				Executable:    log.Meta.Executable,
				Path:          log.Meta.Path,
				Extra:         log.Meta.Extra,
			}
		}

		pbLogs = append(pbLogs, pbLog)
	}

	// Convert outer Meta
	//var convertedMeta *proto.LogMeta
	//if payload.Meta != nil {
	//		convertedMeta = api.ConvertLogMetaToProto(payload.Meta)
	//	}

	req := &proto.LogPayload{
		EndpointId: payload.EndpointID,
		Timestamp:  timestamppb.New(payload.Timestamp),
		Logs:       pbLogs,
	}

	if err := s.stream.Send(req); err != nil {
		return fmt.Errorf("log stream send failed: %w", err)
	}

	return nil
}
