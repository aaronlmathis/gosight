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

	gosightagent "github.com/aaronlmathis/gosight/agent/internal/agent"
	"github.com/aaronlmathis/gosight/agent/internal/bootstrap"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func main() {

	// Bootstrap config loading (flags -> env -> file)
	cfg := bootstrap.LoadAgentConfig()
	fmt.Printf("🔧 About to init logger with level = %s\n", cfg.Logs.LogLevel)

	bootstrap.SetupLogging(cfg)
	utils.Debug("debug logging is active from main.go")

	// Graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		utils.Warn("signal received, shutting down agent...")
		cancel()
	}()

	agent, err := gosightagent.NewAgent(ctx, cfg)
	if err != nil {
		utils.Error("failed to initialize agent: %v", err)
		os.Exit(1)
	}

	agent.Start(ctx)

	<-ctx.Done()
	utils.Info("agent shut down cleanly")
}
