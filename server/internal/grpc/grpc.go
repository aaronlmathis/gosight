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

package grpcserver

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/sys"

	"github.com/aaronlmathis/gosight/server/internal/telemetry"
	"github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type GrpcServer struct {
	Sys           *sys.SystemContext
	LogHandler    *telemetry.LogsHandler
	MetricHandler *telemetry.MetricsHandler
	Listener      net.Listener
	Server        *grpc.Server
}

// NewGRPCServer creates a new gRPC server instance with the provided system context.
// It initializes the server with TLS configuration and registers the metrics and log services.
// The server listens on the address specified in the system context configuration.

func NewGRPCServer(sys *sys.SystemContext) (*GrpcServer, error) {

	// Load TLS for mTLS
	tlsCfg, err := loadTLSConfig(sys.Cfg)
	if err != nil {
		return nil, fmt.Errorf("TLS config failed: %w", err)
	}

	// Create gRPC listener
	listener, err := net.Listen("tcp", sys.Cfg.Server.GRPCAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", sys.Cfg.Server.GRPCAddr, err)
	}

	// Generate credentials from tlsCfg and start gRPC Server
	creds := credentials.NewTLS(tlsCfg)

	server := grpc.NewServer(grpc.Creds(creds),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             1 * time.Minute, // Clients can only ping once per minute
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:                  2 * time.Minute, // Server sends own pings every 2m
			Timeout:               20 * time.Second,
			MaxConnectionIdle:     0,
			MaxConnectionAge:      0,
			MaxConnectionAgeGrace: 0,
		}),
	)
	// Create metric and log handlers
	metricHandler := telemetry.NewMetricsHandler(sys)
	proto.RegisterMetricsServiceServer(server, metricHandler)

	logHandler := telemetry.NewLogsHandler(sys)
	proto.RegisterLogServiceServer(server, logHandler)

	if sys.Cfg.Debug.EnableReflection {
		utils.Info("Enabling gRPC reflection")
		reflection.Register(server)
	}

	return &GrpcServer{
		Sys:           sys,
		LogHandler:    logHandler,
		MetricHandler: metricHandler,
		Listener:      listener,
		Server:        server,
	}, nil

}

// logTLSConfig loads the TLS configuration for the gRPC server.
// It loads the server certificate and key, and sets up client authentication
// if a client CA file is provided. It also verifies the client certificate
// and logs the common name and SHA256 fingerprint of the client certificate.

func loadTLSConfig(cfg *config.Config) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
	if err != nil {
		return nil, err
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return fmt.Errorf("no client cert provided")
			}

			cert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return fmt.Errorf("failed to parse client cert: %w", err)
			}

			// Log CN and fingerprint
			cn := cert.Subject.CommonName
			fingerprint := sha256.Sum256(cert.Raw)
			utils.Info("Agent connected: CN=%s, SHA256 Fingerprint=%s", cn, hex.EncodeToString(fingerprint[:]))

			// Optional: Reject based on CN or SAN here

			return nil
		},
	}

	// Enable mTLS if client CA is provided
	if cfg.TLS.ClientCAFile != "" {
		caCert, err := os.ReadFile(cfg.TLS.ClientCAFile)
		if err != nil {
			return nil, err
		}
		caPool := x509.NewCertPool()
		caPool.AppendCertsFromPEM(caCert)
		tlsCfg.ClientCAs = caPool
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsCfg, nil
}
