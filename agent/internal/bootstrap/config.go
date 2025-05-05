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
	"github.com/aaronlmathis/gosight/shared/utils"
)

func LoadAgentConfig() *config.Config {
	// Flag declarations
	configFlag := flag.String("config", "", "Path to agent config file")
	serverURL := flag.String("server-url", "", "Override server URL")
	interval := flag.Duration("interval", 0, "Override interval (e.g. 5s)")
	host := flag.String("host", "", "Override hostname")

	environment := flag.String("env", "", "Environment (dev, test, prod)")
	logLevel := flag.String("log-level", "", "Log level (debug, info, warn, error)")
	errorLogFile := flag.String("error_log", "", "Path to error log file")
	appLogFile := flag.String("app_log", "", "Path to app log file")
	accessLogFile := flag.String("access_log", "", "Path to access file")
	customTags := flag.String("tags", "", "Comma-separated list of custom tags")

	flag.Parse()

	// Resolve config path
	configPath := resolvePath(*configFlag, "GOSIGHT_AGENT_CONFIG", "config.yaml")
	log.Printf("Loaded config file from: %s", configPath)
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
		cfg.Agent.ServerURL = *serverURL
	}
	if *interval != 0 {
		cfg.Agent.Interval = *interval
	}
	if *host != "" {
		cfg.Agent.HostOverride = *host
	}
	if *environment != "" {
		cfg.Agent.Environment = *environment
	}

	if *logLevel != "" {
		cfg.Logs.LogLevel = *logLevel
	}
	if *appLogFile != "" {
		cfg.Logs.AppLogFile = *appLogFile
	}
	if *errorLogFile != "" {
		cfg.Logs.ErrorLogFile = *errorLogFile
	}
	if *accessLogFile != "" {
		cfg.Logs.AccessLogFile = *errorLogFile
	}
	if *customTags != "" {
		cfg.CustomTags = utils.ParseTagString(*customTags)
	}

	return cfg
}

func resolvePath(flagVal, envVar, fallback string) string {
	if flagVal != "" {
		return absPath(flagVal)
	}
	if val := os.Getenv(envVar); val != "" {
		utils.Debug("Using %s from environment variable: %s", envVar, val)
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
