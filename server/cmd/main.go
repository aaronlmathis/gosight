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

// File: server/cmd/main.go
package main

import (
	"fmt"

	"github.com/aaronlmathis/gosight/server/internal/bootstrap"
	grpcserver "github.com/aaronlmathis/gosight/server/internal/grpc"
	httpserver "github.com/aaronlmathis/gosight/server/internal/http"
	"github.com/aaronlmathis/gosight/server/internal/http/websocket"
	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func main() {

	// Bootstrap config loading (flags -> env -> file)
	cfg := bootstrap.LoadServerConfig()
	fmt.Printf("🔧 About to init logger with level = %s\n", cfg.Logs.LogLevel)

	// Initialize logging
	bootstrap.SetupLogging(cfg)

	// Initialize the websocket hub
	wsHub := websocket.NewHub()
	go func() {
		utils.Info("🧬 Starting WebSocket hub...")
		wsHub.Run() // no error returned, but safe to log around
	}()

	// Init metric store
	metricStore, err := bootstrap.InitMetricStore(cfg)
	utils.Must("Metric store", err)

	// Initialize agent tracker
	agentTracker, err := bootstrap.InitAgentTracker(cfg.Server.Environment)
	utils.Must("Agent tracker", err)

	// Initialize metric index
	metricIndex, err := bootstrap.InitMetricIndex()
	utils.Must("Metric index", err)

	// Initialize user store
	userStore, err := bootstrap.InitUserStore(cfg)
	utils.Must("User store", err)

	// Initialize meta tracker
	metaTracker := metastore.NewMetaTracker()

	// Initialize auth
	authProviders, err := httpserver.InitAuth(cfg, userStore)
	utils.Must("Auth providers", err)

	// Start HTTP server for admin console/api
	srv := httpserver.NewServer(agentTracker, authProviders, cfg, metaTracker, metricIndex, metricStore, userStore, wsHub)

	go func() {
		if err := srv.Start(); err != nil {
			utils.Fatal("HTTP server failed: %v", err)
		} else {
			utils.Info("🌐 HTTP server started successfully")
		}
	}()

	grpcServer, listener, err := grpcserver.NewGRPCServer(cfg, metricStore, agentTracker, metricIndex, metaTracker, wsHub)
	if err != nil {
		utils.Fatal("Failed to start gRPC server: %v", err)
	} else {
		utils.Info("🚀 GoSight server listening on %s", cfg.Server.GRPCAddr)
		if err := grpcServer.Serve(listener); err != nil {
			utils.Fatal("Failed to serve: %v", err)
		}
	}
}
