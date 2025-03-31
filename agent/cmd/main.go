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
along with LeetScraper. If not, see https://www.gnu.org/licenses/.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/config"
)

func main() {

	// Flag declarations
	configFlag := flag.String("config", "config.yaml", "Path to agent config file")
	serverURL := flag.String("server-url", "", "Override server URL")
	interval := flag.Duration("interval", 0, "Override interval (e.g. 5s)")
	host := flag.String("host", "", "Override hostname")
	metrics := flag.String("metrics", "", "Comma-separated list of enabled metrics")

	// Parse all flags first
	flag.Parse()

	// Resolve config path from flag, env var, or default
	resolvedPath := resolveConfigPath(*configFlag, "AGENT_CONFIG", "config.yaml")

	// Create default if missing
	if err := config.EnsureDefaultConfig(resolvedPath); err != nil {
		log.Fatalf("Could not create default config: %v", err)
	}

	// Load config
	cfg, err := config.LoadConfig(resolvedPath)
	if err != nil {
		log.Fatalf("Failed to load agent config: %v", err)
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

	fmt.Println("Effective Agent Config:")
	fmt.Printf("  Server URL: %s\n", cfg.ServerURL)
	fmt.Printf("  Interval: %v\n", cfg.Interval)
	fmt.Printf("  Host: %s\n", cfg.HostOverride)
	fmt.Printf("  Metrics Enabled: %v\n", cfg.MetricsEnabled)

	time.Sleep(1 * time.Second)
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
		log.Fatalf("Failed to resolve path: %v", err)
	}
	return abs
}
