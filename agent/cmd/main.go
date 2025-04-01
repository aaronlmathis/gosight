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
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/agent/internal/runner"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func main() {

	// Flag declarations
	configFlag := flag.String("config", "config.yaml", "Path to agent config file")
	serverURL := flag.String("server-url", "", "Override server URL")
	interval := flag.Duration("interval", 0, "Override interval (e.g. 5s)")
	host := flag.String("host", "", "Override hostname")
	metrics := flag.String("metrics", "", "Comma-separated list of enabled metrics")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	logFile := flag.String("log-file", "agent.log", "Path to log file")

	// Parse all flags first
	flag.Parse()

	// Resolve config path from flag, env var, or default
	resolvedPath := resolveConfigPath(*configFlag, "AGENT_CONFIG", "config.yaml")

	// Create default if missing
	if err := config.EnsureDefaultConfig(resolvedPath); err != nil {
		fmt.Printf("Could not create default config: %v", err)
		os.Exit(1)
	}

	// Load config
	cfg, err := config.LoadConfig(resolvedPath)
	if err != nil {
		fmt.Printf("Failed to load agent config: %v", err)
		os.Exit(1)
	}

	// Apply ENV var overrides
	config.ApplyEnvOverrides(cfg)

	// Apply CLI flag overrides (highest priority)
	if *serverURL != "" {
		cfg.ServerURL = *serverURL
	}
	if *interval != 0 {
		cfg.Interval = *interval
	}
	if *host != "" {
		cfg.HostOverride = *host
	}
	if *metrics != "" {
		cfg.MetricsEnabled = config.SplitCSV(*metrics)
	}
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}
	if *logFile != "" {
		cfg.LogFile = *logFile
	}

	if err := utils.InitLogger(cfg.LogFile, cfg.LogLevel); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

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

	fmt.Println("Effective Agent Config:")
	fmt.Printf("  Server URL: %s\n", cfg.ServerURL)
	fmt.Printf("  Interval: %v\n", cfg.Interval)
	fmt.Printf("  Host: %s\n", cfg.HostOverride)
	fmt.Printf("  Metrics Enabled: %v\n", cfg.MetricsEnabled)
	fmt.Printf("  Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("  Log File: %s\n", cfg.LogFile)

	// start streaming agent
	runner.RunAgent(ctx, cfg)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func resolveConfigPath(flagVal, envVar, fallback string) string {
	if flagVal != "" {
		return mustAbs(flagVal)
	}
	if val := os.Getenv(envVar); val != "" {
		return mustAbs(val)
	}
	return mustAbs(fallback)
}

func mustAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		fmt.Printf("Failed to resolve path: %v", err)
		os.Exit(1)
	}
	return abs
}