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
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aaronlmathis/gosight/server/internal/bootstrap"
	grpcserver "github.com/aaronlmathis/gosight/server/internal/grpc"
	httpserver "github.com/aaronlmathis/gosight/server/internal/http"
	"github.com/aaronlmathis/gosight/server/internal/http/websocket"
	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/aaronlmathis/gosight/shared/utils"
)

var Version = "dev" // default
// go build -ldflags "-X main.Version=0.3.2" -o gosight-agent ./cmd/agent

func main() {

	// Graceful Shutdown Context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		sig := <-sigCh
		utils.Warn("Received signal: %s", sig)

		cancel() // cancels context for all background loops

		// Optional: close server listeners explicitly here
	}()

	// Bootstrap config loading (flags -> env -> file)
	cfg := bootstrap.LoadServerConfig()
	fmt.Printf("About to init logger with level = %s\n", cfg.Logs.LogLevel)

	// Initialize logging
	bootstrap.SetupLogging(cfg)

	// Initialize the websocket hub
	wsHub := websocket.NewHub()
	go func() {
		utils.Info("Starting WebSocket hub...")
		wsHub.Run() // no error returned, but safe to log around
	}()

	// Init metric store
	metricStore, err := bootstrap.InitMetricStore(ctx, cfg)
	utils.Must("Metric store", err)

	// Initialize user store
	dataStore, err := bootstrap.InitDataStore(cfg)
	utils.Must("Data store", err)

	// Initialize agent tracker
	agentTracker, err := bootstrap.InitAgentTracker(ctx, cfg.Server.Environment, dataStore)
	utils.Must("Agent tracker", err)

	// Initialize metric index
	metricIndex, err := bootstrap.InitMetricIndex()
	utils.Must("Metric index", err)

	// Initialize meta tracker
	metaTracker := metastore.NewMetaTracker()

	// Initialize user store
	userStore, err := bootstrap.InitUserStore(cfg)
	utils.Must("User store", err)

	// Initialize auth
	authProviders, err := httpserver.InitAuth(cfg, userStore)
	utils.Must("Auth providers", err)

	// Start HTTP server for admin console/api
	srv := httpserver.NewServer(ctx, agentTracker, authProviders, cfg, metaTracker, metricIndex, metricStore, userStore, wsHub)

	go func() {
		if err := srv.Start(); err != nil {
			utils.Fatal("HTTP server failed: %v", err)
		} else {
			utils.Info("HTTP server started successfully")
		}
	}()

	grpcServer, listener, err := grpcserver.NewGRPCServer(ctx, cfg, metricStore, agentTracker, metricIndex, metaTracker, wsHub)
	if err != nil {
		utils.Fatal("Failed to start gRPC server: %v", err)
	} else {
		go func() {
			utils.Info("GoSight server listening on %s", cfg.Server.GRPCAddr)
			if err := grpcServer.Serve(listener); err != nil {
				utils.Fatal("Failed to serve gRPC: %v", err)
			}
		}()
	}

	<-ctx.Done()
	utils.Info("🧹 Shutting down GoSight...")

	grpcServer.GracefulStop()
	_ = srv.Shutdown(ctx)
	_ = metricStore.Close()
}
