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
	"fmt"
	"log"
	"time"

	"github.com/aaronlmathis/gosight/internal/collector"
	"github.com/aaronlmathis/gosight/internal/config"
	"github.com/aaronlmathis/gosight/internal/exporter"
)

func main() {

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Create the store to hold latest metrics
	store := collector.NewMetricStore()

	// Create the results channel to receive updates from collectors
	results := make(chan collector.MetricResult)

	// Register collectors
	collectors := []collector.MetricCollector{
		&collector.CPUCollector{},
		&collector.MemoryCollector{},
		collector.NewDiskCollector(),
		&collector.NetCollector{},
	}

	// Start collectors in background goroutines
	interval := time.Duration(cfg.Metrics.IntervalSeconds) * time.Second
	collector.StartCollectors(collectors, interval, results)
	collector.StartCollectors(collectors, 2*time.Second, results)

	// Handle results and update store
	go func() {
		for result := range results {
			if result.Err != nil {
				log.Printf("[%s] error: %v", result.Name, result.Err)
				continue
			}
			store.Update(result.Data)
			log.Printf("[%s] updated: %v", result.Name, result.Data)
		}
	}()
	go exporter.StartHTTPServer(
		fmt.Sprintf(":%d", cfg.Server.Port),
		store,
		cfg.Thresholds,
		cfg.Exporters.Prometheus,
		cfg.Exporters.Dashboard,
	)

	// Block main forever (until we implement graceful shutdown)
	select {}
}
