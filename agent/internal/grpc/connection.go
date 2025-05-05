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
		// Optionally, add a check here to see if the existing connection is still healthy
		// using conn.GetState() or a simple RPC call.
		return conn, nil
	}

	tlsCfg, err := agentutils.LoadTLSConfig(cfg)
	if err != nil {
		return nil, err
	}

	// Options previously passed to DialContext
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                2 * time.Minute,
			Timeout:             20 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithInitialWindowSize(64 * 1024 * 1024),
		grpc.WithInitialConnWindowSize(128 * 1024 * 1024),
		grpc.WithReadBufferSize(8 * 1024 * 1024),
		grpc.WithWriteBufferSize(8 * 1024 * 1024),
		grpc.WithDefaultCallOptions(
			grpc.UseCompressor(gzip.Name),
			grpc.MaxCallRecvMsgSize(32 * 1024 * 1024),
			grpc.MaxCallSendMsgSize(32 * 1024 * 1024),
		),
		// If you were relying on DialContext's implicit "passthrough" resolver for non-standard
		// targets (like in-memory listeners for testing), you might need to explicitly specify
		// the "passthrough" scheme with NewClient:
		// grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"passthrough":{}}]}`),
		// For standard "host:port" addresses that should use DNS resolution, no extra option is needed
		// as "dns" is the default for NewClient.
	}

	// Use grpc.NewClient instead of grpc.DialContext
	c, err := grpc.NewClient(cfg.Agent.ServerURL, opts...)
	if err != nil {
		return nil, err
	}

	// Note: NewClient returns a ClientConn immediately.
	// It does NOT block until the connection is established, unlike DialContext with WithBlock.
	// The connection happens in the background. RPC calls made on the connection will block
	// until the connection is ready or the call's context times out.
	// If you need to wait for the connection to be established before returning,
	// you would need to implement a mechanism to wait for the connection state to become READY,
	// possibly using conn.WaitForStateChange(). However, for most use cases, this isn't necessary
	// as the RPC calls themselves handle the waiting.

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
