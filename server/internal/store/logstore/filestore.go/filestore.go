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

// gosight/server/internal/store/logstore/filestore.go
// Defines filestore

package filestore

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type FileStore struct {
	Config   *config.Config
	Dir      string
	queue    chan []model.LogPayload
	incoming chan []model.LogPayload
	wg       sync.WaitGroup
	client   *http.Client
	ctx      context.Context

	// batching config
	batchSize     int
	batchTimeout  time.Duration
	batchRetry    int
	batchInterval time.Duration
}

func NewFileStore(ctx context.Context, cfg *config.Config) *FileStore {

	workers := cfg.LogStore.Workers
	utils.Info("NewFileStore received workers=%d", workers)
	if _, err := os.Stat(cfg.LogStore.Dir); os.IsNotExist(err) {
		if err := os.Mkdir(cfg.LogStore.Dir, 0755); err != nil {
			utils.Error("Failed to create log store directory: %v", err)
		}
	}
	filestore := &FileStore{
		Config:        cfg,
		Dir:           cfg.LogStore.Dir,
		queue:         make(chan []model.LogPayload, cfg.LogStore.BatchSize),
		incoming:      make(chan []model.LogPayload, cfg.LogStore.BatchSize),
		client:        &http.Client{Timeout: 10 * time.Second},
		ctx:           ctx,
		batchSize:     cfg.LogStore.BatchSize,
		batchTimeout:  time.Duration(cfg.LogStore.BatchTimeout) * time.Millisecond,
		batchRetry:    cfg.LogStore.BatchRetry,
		batchInterval: time.Duration(cfg.LogStore.BatchInterval) * time.Millisecond,
	}

	if workers == 0 {
		utils.Warn("JSON File Store called with 0 workers!")
	} else {
		utils.Debug("Spawning %d workers now...", workers)
	}

	// Start up workers
	for i := 0; i < workers; i++ {
		filestore.wg.Add(1)

		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					utils.Error("Worker #%d panicked: %v", id, r)
				}
			}()
			utils.Info("Started worker #%d", id)
			filestore.worker()
		}(i + 1)
	}

	go filestore.collectorLoop()

	return filestore
}

func (v *FileStore) Write(logs []model.LogPayload, streamCtx context.Context) error {
	utils.Debug(" filestore.Write received: %d metrics (store addr: %p)", totalLogCount(logs), v)

	select {
	case v.incoming <- logs:
		//utils.Debug("Write enqueued %d metrics", totalMetricCount(metrics))
		return nil
	default:
		utils.Warn("Incoming buffer full: dropping metrics")
		return fmt.Errorf("incoming buffer full")
	}
}

func (v *FileStore) collectorLoop() {
	utils.Info("filestore collectorLoop started")
	ticker := time.NewTicker(v.batchTimeout)
	defer ticker.Stop()

	//utils.Info("⏱batchTimeout raw = %v\n", v.batchTimeout)
	//utils.Debug("collectorLoop started with timeout: %s", v.batchTimeout)

	var pending []model.LogPayload

	for {
		select {
		case <-v.ctx.Done():
			utils.Debug("LogStore collector loop exiting")
			if len(pending) > 0 {
				v.enqueue(pending)
			}
			return

		case batch := <-v.incoming:
			//total := totalMetricCount(batch)

			pending = append(pending, batch...)
			currentTotal := totalLogCount(pending)
			utils.Debug(" Received payload with %d metrics", currentTotal)
			//utils.Debug(" Total metrics pending: %d", currentTotal)

			if currentTotal >= v.batchSize {
				//utils.Info("📦 Batch size reached: %d metrics, flushing now", currentTotal)
				v.enqueue(pending)
				pending = nil
			}

		case <-ticker.C:
			currentTotal := totalLogCount(pending)
			//utils.Debug(" Timeout ticked. Pending payloads: %d, metrics: %d", len(pending), currentTotal)

			if currentTotal > 0 {
				//utils.Info(" Timeout flush triggered for %d metrics", currentTotal)
				v.enqueue(pending)
				pending = nil
			}
		}
	}
}

func (v *FileStore) worker() {
	defer v.wg.Done()
	for {
		utils.Debug("Filestore Worker waiting for batch...")

		select {

		case batch := <-v.queue:
			utils.Debug("Filestore Worker received batch with %d payloads / %d logs", len(batch), totalLogCount(batch))
			v.flush(batch)
		case <-v.ctx.Done():
			utils.Debug("Logstore collector loop exiting")

			return
		}
	}
}

func (v *FileStore) enqueue(batch []model.LogPayload) {
	//utils.Debug("Enqueue called with %d payloads / %d metrics",		len(batch), totalMetricCount(batch))
	select {
	case v.queue <- batch:
	default:
		utils.Warn("Worker queue full: dropping batch of %d metrics", len(batch))
	}
}

func (v *FileStore) flush(batch []model.LogPayload) {

	for _, payload := range batch {
		if len(payload.Logs) == 0 {
			continue
		}

		// Construct compressed output path
		timestamp := time.Now().UTC().Format("20060102T150405Z")
		filename := fmt.Sprintf("logs_%s_%s.json.gz", payload.EndpointID, timestamp)
		path := filepath.Join(v.Dir, filename)

		// Open file
		f, err := os.Create(path)
		if err != nil {
			utils.Error("Failed to create file %s: %v", path, err)
			continue
		}

		// Gzip writer
		gz := gzip.NewWriter(f)
		enc := json.NewEncoder(gz)
		enc.SetIndent("", "  ")

		if err := enc.Encode(payload); err != nil {
			utils.Error("Failed to encode log payload: %v", err)
		}

		_ = gz.Close()
		_ = f.Close()

		utils.Debug("Wrote %d logs to %s", len(payload.Logs), filename)
	}
}

func (v *FileStore) Close() error {
	utils.Info("Waiting for VictoriaStore workers to finish...")
	v.wg.Wait()
	utils.Info("VictoriaStore shutdown complete")
	return nil
}

// HELPERS

func totalLogCount(payloads []model.LogPayload) int {
	count := 0
	for _, p := range payloads {
		count += len(p.Logs)
	}
	return count
}
