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

	"github.com/aaronlmathis/gosight/server/internal/config"
)

func main() {

	// CLI Flag declarations
	configFlag := flag.String("config", "config.yaml", "Path to agent config file")
	listen := flag.String("listen", "", "Override listen address")
	storage := flag.String("storage", "", "Override storage engine")
	dbPath := flag.String("db-path", "", "Override database path")

	// Parse all flags first
	flag.Parse()

	// Resolve config path from flag, env var, or default
	resolvedPath := resolveConfigPath(*configFlag, "SERVER_CONFIG", "config.yaml")

	// Create default if missing
	if err := config.EnsureDefaultConfig(resolvedPath); err != nil {
		log.Fatalf("Could not create default config: %v", err)
	}

	// Load YAML config
	cfg, err := config.LoadConfig(resolvedPath)
	if err != nil {
		log.Fatalf("Failed to load server config: %v", err)
	}

	// Apply ENV var overrides
	config.ApplyEnvOverrides(cfg)

	// Apply CLI flag overrides (highest priority)
	if *listen != "" {
		cfg.ListenAddr = *listen
	}
	if *storage != "" {
		cfg.StorageEngine = *storage
	}
	if *dbPath != "" {
		cfg.DatabasePath = *dbPath
	}

	// Start the server
	fmt.Printf("GoBright Server listening on %s (storage: %s, DB: %s)\n",
		cfg.ListenAddr, cfg.StorageEngine, cfg.DatabasePath)

	//http.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintln(w, "OK: metrics received (TODO)")
	//})

	//log.Fatal(http.ListenAndServe(cfg.ListenAddr, nil))
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
