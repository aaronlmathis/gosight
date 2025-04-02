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

// File: agent/internal/bootstrap/config.go
// Loads ENV, FLAG, Configs

package bootstrap

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/aaronlmathis/gosight/agent/internal/config"
)

func LoadServerConfig() *config.AgentConfig {
	// Flag declarations
	configFlag := flag.String("config", "", "Path to agent config file")
	serverURL := flag.String("server-url", "", "Override server URL")
	interval := flag.Duration("interval", 0, "Override interval (e.g. 5s)")
	host := flag.String("host", "", "Override hostname")
	metrics := flag.String("metrics", "", "Comma-separated list of enabled metrics")
	logLevel := flag.String("log-level", "", "Log level (debug, info, warn, error)")
	logFile := flag.String("log-file", "", "Path to log file")

	flag.Parse()

	// Resolve config path
	configPath := resolvePath(*configFlag, "AGENT_CONFIG", "config.yaml")

	if err := config.EnsureDefaultConfig(configPath); err != nil {
		log.Fatalf("Could not create default config: %v", err)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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

	return cfg
}

func resolvePath(flagVal, envVar, fallback string) string {
	if flagVal != "" {
		return absPath(flagVal)
	}
	if val := os.Getenv(envVar); val != "" {
		return absPath(val)
	}
	return absPath(fallback)
}

func absPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Failed to resolve path: %v", err)
	}
	return abs
}
