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

// gosight/server/internal/server/
// grpc.go - gRPC server setup and initialization

package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	"github.com/aaronlmathis/gosight/server/internal/api"
	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

func NewGRPCServer(cfg *config.ServerConfig, store store.MetricStore) (*grpc.Server, net.Listener, error) {
	tlsCfg, err := loadTLSConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("TLS config failed: %w", err)
	}
	listener, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on %s: %w", cfg.ListenAddr, err)
	}

	creds := credentials.NewTLS(tlsCfg)
	server := grpc.NewServer(grpc.Creds(creds))

	handler := api.NewMetricsHandler(store)
	proto.RegisterMetricsServiceServer(server, handler)

	utils.Debug("📨 NewGRPCServer received store at: %p", store)

	if cfg.Debug.EnableReflection {
		utils.Info("Enabling gRPC reflection")
		reflection.Register(server)
	}

	return server, listener, nil
}

func loadTLSConfig(cfg *config.ServerConfig) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
	if err != nil {
		return nil, err
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
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
