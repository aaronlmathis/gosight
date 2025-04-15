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
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"net"
	"os"

	"github.com/aaronlmathis/gosight/server/internal/api"
	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/http/websocket"
	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

func NewGRPCServer(ctx context.Context, cfg *config.Config, store store.MetricStore, tracker *store.AgentTracker, metricIndex *store.MetricIndex, metaTracker *metastore.MetaTracker, ws *websocket.Hub) (*grpc.Server, net.Listener, error) {
	tlsCfg, err := loadTLSConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("TLS config failed: %w", err)
	}
	listener, err := net.Listen("tcp", cfg.Server.GRPCAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on %s: %w", cfg.Server.GRPCAddr, err)
	}

	creds := credentials.NewTLS(tlsCfg)
	server := grpc.NewServer(grpc.Creds(creds))

	handler := api.NewMetricsHandler(store, tracker, metricIndex, metaTracker, ws)
	proto.RegisterMetricsServiceServer(server, handler)

	utils.Debug("📨 NewGRPCServer received store at: %p", store)

	if cfg.Debug.EnableReflection {
		utils.Info("Enabling gRPC reflection")
		reflection.Register(server)
	}

	return server, listener, nil
}

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
			utils.Info("🔐 Agent connected: CN=%s, SHA256 Fingerprint=%s", cn, hex.EncodeToString(fingerprint[:]))

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
