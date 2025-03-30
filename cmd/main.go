package main

import (
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

	// Container Support
	if cfg.Containers.Enabled {
		if cfg.Containers.Runtime == "podman" {
			pc := collector.NewPodmanCollector(cfg.Containers.Socket)
			collectors = append(collectors, pc)
			log.Println("[container] using native PodmanCollector")
		}
		// TODO: Add Docker support
	}

	// Start collectors
	interval := time.Duration(cfg.Metrics.IntervalSeconds) * time.Second
	collector.StartCollectors(collectors, interval, results)

	// Handle results and update store
	go func() {
		for result := range results {
			if result.Err != nil {
				log.Printf("[%s] error: %v", result.Name, result.Err)
				continue
			}
			store.Update(result.Data, result.Meta)
			log.Printf("[%s] updated: %v", result.Name, result.Data)
		}
	}()

	// Start HTTP server
	go exporter.StartHTTPServer(cfg, store)

	// Block forever
	select {}
}
