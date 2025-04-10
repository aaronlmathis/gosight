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

// cmd/main.go - main entry point for agent.

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aaronlmathis/gosight/agent/internal/bootstrap"
	"github.com/aaronlmathis/gosight/agent/internal/runner"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func main() {

	// Bootstrap config loading (flags -> env -> file)
	cfg := bootstrap.LoadAgentConfig()
	fmt.Printf("🔧 About to init logger with level = %s\n", cfg.Logs.LogLevel)
	bootstrap.SetupLogging(cfg)
	utils.Debug("✅ Debug logging is active from main.go")
	utils.Info("📡 Connecting to server at: %s", cfg.Agent.ServerURL)
	// Graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		utils.Warn("🔌 Signal received, shutting down agent...")
		cancel()
	}()

	// start streaming agent
	runner.RunAgent(ctx, cfg)
}
