package grpcconn

import (
	"context"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
