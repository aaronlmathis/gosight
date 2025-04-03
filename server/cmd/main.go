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
	httpserver "github.com/aaronlmathis/gosight/server/internal/http"
	"github.com/aaronlmathis/gosight/server/internal/server"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func main() {

	// Bootstrap config loading (flags -> env -> file)
	cfg := bootstrap.LoadServerConfig()
	fmt.Printf("🔧 About to init logger with level = %s\n", cfg.Server.LogLevel)
	// Initialize logging
	bootstrap.SetupLogging(cfg)

	// Init metric store

	metricStore, err := bootstrap.InitMetricStore(cfg)
	if err != nil {
		utils.Fatal("Metric store init failed: %v", err)
	}

	agentTracker, err := bootstrap.InitAgentTracker(cfg.Server.Environment)
	if err != nil {
		utils.Fatal("Agent tracker init failed: %v", err)
	}
	// Start HTTP server for admin console/api

	go httpserver.StartHTTPServer(cfg, agentTracker)

	grpcServer, listener, err := server.NewGRPCServer(cfg, metricStore, agentTracker)
	if err != nil {
		utils.Fatal("Failed to start gRPC server: %v", err)
	}
	utils.Info("🚀 GoSight server listening on %s", cfg.Server.GRPCAddr)
	if err := grpcServer.Serve(listener); err != nil {
		utils.Fatal("Failed to serve: %v", err)
	}

}
