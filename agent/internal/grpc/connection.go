package grpcconn

import (
	"context"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
)

var (
	conn   *grpc.ClientConn
	connMu sync.Mutex
)

// GetGRPCConn returns the singleton ClientConn
func GetGRPCConn(ctx context.Context, cfg *config.Config) (*grpc.ClientConn, error) {
	connMu.Lock()
	defer connMu.Unlock()

	if conn != nil {
		return conn, nil
	}

	tlsCfg, err := agentutils.LoadTLSConfig(cfg)
	if err != nil {
		return nil, err
	}

	c, err := grpc.DialContext(
		ctx,
		cfg.Agent.ServerURL,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)),

		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                2 * time.Minute,
			Timeout:             20 * time.Second,
			PermitWithoutStream: true,
		}),

		grpc.WithInitialWindowSize(64*1024*1024),
		grpc.WithInitialConnWindowSize(128*1024*1024),

		grpc.WithReadBufferSize(8*1024*1024),
		grpc.WithWriteBufferSize(8*1024*1024),

		grpc.WithDefaultCallOptions(
			grpc.UseCompressor(gzip.Name),
			grpc.MaxCallRecvMsgSize(32*1024*1024),
			grpc.MaxCallSendMsgSize(32*1024*1024),
		),
	)

	if err != nil {
		return nil, err
	}

	conn = c
	return conn, nil
}

// CloseGRPCConn closes the connection (for shutdown)
func CloseGRPCConn() error {
	connMu.Lock()
	defer connMu.Unlock()
	if conn != nil {
		err := conn.Close()
		conn = nil
		return err
	}
	return nil
}
