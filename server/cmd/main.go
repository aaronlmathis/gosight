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
	"os"
	"os/signal"
	"syscall"

	"github.com/aaronlmathis/gosight/server/internal/bootstrap"
	grpcserver "github.com/aaronlmathis/gosight/server/internal/grpc"
	httpserver "github.com/aaronlmathis/gosight/server/internal/http"
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

		cancel()

	}()

	// Init System Context for Gosight
	sys, err := bootstrap.InitGoSight(ctx)
	utils.Must("System Context", err)

	// Start HTTP server for admin console/api
	srv := httpserver.NewServer(sys)

	go func() {
		if err := srv.Start(); err != nil {
			utils.Fatal("HTTP server failed: %v", err)
		} else {
			utils.Info("HTTP server started successfully")
		}
	}()

	grpcServer, err := grpcserver.NewGRPCServer(sys)
	if err != nil {
		utils.Fatal("Failed to start gRPC server: %v", err)
	} else {
		go func() {
			utils.Info("GoSight server listening on %s", sys.Cfg.Server.GRPCAddr)
			if err := grpcServer.Server.Serve(grpcServer.Listener); err != nil {
				utils.Fatal("Failed to serve gRPC: %v", err)
			}
		}()
	}

	<-ctx.Done()
	utils.Info("🧹 Shutting down GoSight...")

	grpcServer.Server.GracefulStop()
	if err := srv.Shutdown(); err != nil {
		utils.Warn("Failed to shutdown HTTP server: %v", err)
	}
	if err := sys.Stores.Metrics.Close(); err != nil {
		utils.Warn("Failed to close metric store: %v", err)
	}
	if err := sys.Stores.Data.Close(); err != nil {
		utils.Warn("Failed to close datastore: %v", err)
	}
	if err := sys.Stores.Users.Close(); err != nil {
		utils.Warn("Failed to close userstore: %v", err)
	}
}

/* Proposed organization of the code:

type StoreModule struct {
	Metrics store.MetricStore
	Logs    logstore.LogStore
	Users   userstore.UserStore
	Data    datastore.DataStore
	Events  eventstore.Store
}

type TelemetryModule struct {
	Index     *store.MetricIndex
	Meta      *metastore.MetaTracker
	Evaluator *rules.Evaluator
	Alerts    *alerts.Manager
}
type SystemContext struct {
	Ctx     context.Context
	Cfg     *config.Config
	Agents  *store.AgentTracker
	Web     *websocket.Hub
	Auth    map[string]auth.AuthProvider
	Stores  *StoreModule
	Tele    *TelemetryModule
}

stores := &StoreModule{
	Metrics: metricStore,
	Logs:    logStore,
	Users:   userStore,
	Data:    dataStore,
	Events:  eventStore,
}

telemetry := &TelemetryModule{
	Index:     metricIndex,
	Meta:      metaTracker,
	Evaluator: evaluator,
	Alerts:    alertMgr,
}

sys := &SystemContext{
	Ctx:     context.Background(),
	Cfg:     cfg,
	Agents:  agentTracker,
	Web:     wsHub,
	Auth:    authProviders,
	Stores:  stores,
	Tele:    telemetry,
}

srv := httpserver.NewServer(sys)
grpcSrv, listener := grpcserver.NewServer(sys)

Internally:
metricStore := sys.Stores.Metrics
alertMgr := sys.Tele.Alerts
cfg := sys.Cfg
*/
