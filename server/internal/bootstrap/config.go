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

// File: server/internal/bootstrap/config.go
// Loads ENV, FLAG, Configs

package bootstrap

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/aaronlmathis/gosight/server/internal/config"
)

func LoadServerConfig() *config.Config {
	// CLI flags
	configFlag := flag.String("config", "", "Path to server config file")
	grpcFlag := flag.String("grpc", "", "Overrid gRPC servere listen address")
	httpFlag := flag.String("http", "", "Overrid http servere listen address")
	envFlag := flag.String("env", "", "Environment (dev / test / prod)")
	logLevelFlag := flag.String("log-level", "", "Log level (debug, info, warn, error)")
	appLogFileFlag := flag.String("applog", "", "Path to log file")
	errorLogFileFlag := flag.String("errorlog", "", "Path to log file")
	flag.Parse()

	// Resolve config path
	configPath := resolvePath(*configFlag, "SERVER_CONFIG", "config.yaml")

	if err := config.EnsureDefaultConfig(configPath); err != nil {
		log.Fatalf("Could not create default config: %v", err)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	config.ApplyEnvOverrides(cfg)

	// Apply CLI flag overrides (highest priority)
	if *grpcFlag != "" {
		cfg.Server.GRPCAddr = *grpcFlag
	}
	if *httpFlag != "" {
		cfg.Server.HTTPAddr = *httpFlag
	}
	if *envFlag != "" {
		cfg.Server.Environment = *envFlag
	}
	if *logLevelFlag != "" {
		cfg.Logs.LogLevel = *logLevelFlag
	}
	if *appLogFileFlag != "" {
		cfg.Logs.AppLogFile = *appLogFileFlag
	}
	if *errorLogFileFlag != "" {
		cfg.Logs.AppLogFile = *errorLogFileFlag
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
