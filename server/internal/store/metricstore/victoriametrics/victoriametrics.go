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
along with GoBright. If not, see https://www.gnu.org/licenses/.
*/

// server/internal/store/victoriametrics.go

package victoriametricstore

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/store/metricindex"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type VictoriaStore struct {
	url      string
	queue    chan []model.MetricPayload
	incoming chan []model.MetricPayload
	wg       sync.WaitGroup
	client   *http.Client
	ctx      context.Context

	// batching config
	batchSize     int
	batchTimeout  time.Duration
	batchRetry    int
	batchInterval time.Duration
	MetricIndex   *metricindex.MetricIndex
}

func NewVictoriaStore(ctx context.Context, url string, workers, queueSize, batchSize, timeoutMS, retry, retryIntervalMS int, metricIndex *metricindex.MetricIndex) *VictoriaStore {
	utils.Info("NewVictoriaStore received workers=%d", workers)
	store := &VictoriaStore{
		url:           url,
		queue:         make(chan []model.MetricPayload, queueSize),
		incoming:      make(chan []model.MetricPayload, queueSize),
		client:        &http.Client{Timeout: 10 * time.Second},
		ctx:           ctx,
		batchSize:     batchSize,
		batchTimeout:  time.Duration(timeoutMS) * time.Millisecond,
		batchRetry:    retry,
		batchInterval: time.Duration(retryIntervalMS) * time.Millisecond,
		MetricIndex:   metricIndex,
	}
	if workers == 0 {
		utils.Warn("VictoriaStore called with 0 workers!")
	} else {
		utils.Debug("Spawning %d workers now...", workers)
	}

	for i := 0; i < workers; i++ {
		store.wg.Add(1)

		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					utils.Error("Worker #%d panicked: %v", id, r)
				}
			}()
			utils.Info("🧵 Started worker #%d", id)
			store.worker()
		}(i + 1)
	}

	go store.collectorLoop()

	utils.Info("VictoriaStore initialized with %d workers", workers)
	utils.Debug("NewVictoriaStore created at address: %p", store)

	return store
}

func (v *VictoriaStore) Write(metrics []model.MetricPayload) error {
	//utils.Debug(" store.Write received: %d metrics (store addr: %p)", totalMetricCount(metrics), v)

	select {
	case v.incoming <- metrics:
		//utils.Debug("Write enqueued %d metrics", totalMetricCount(metrics))
		return nil
	default:
		utils.Warn("Incoming buffer full: dropping metrics")
		return fmt.Errorf("incoming buffer full")
	}
}

func (v *VictoriaStore) collectorLoop() {
	utils.Info("collectorLoop started")
	ticker := time.NewTicker(v.batchTimeout)
	defer ticker.Stop()

	//utils.Info("atchTimeout raw = %v\n", v.batchTimeout)
	//utils.Debug("collectorLoop started with timeout: %s", v.batchTimeout)

	var pending []model.MetricPayload

	for {
		select {
		case <-v.ctx.Done():
			utils.Debug("VictoriaStore collector loop exiting")
			if len(pending) > 0 {
				v.enqueue(pending)
			}
			return

		case batch := <-v.incoming:
			//total := totalMetricCount(batch)
			//utils.Debug(" Received payload with %d metrics", total)
			pending = append(pending, batch...)
			currentTotal := totalMetricCount(pending)
			//utils.Debug(" Total metrics pending: %d", currentTotal)

			if currentTotal >= v.batchSize {
				//utils.Info(" Batch size reached: %d metrics, flushing now", currentTotal)
				v.enqueue(pending)
				pending = nil
			}

		case <-ticker.C:
			currentTotal := totalMetricCount(pending)
			//utils.Debug(" Timeout ticked. Pending payloads: %d, metrics: %d", len(pending), currentTotal)

			if currentTotal > 0 {
				//utils.Info(" Timeout flush triggered for %d metrics", currentTotal)
				v.enqueue(pending)
				pending = nil
			}
		}
	}
}

func (v *VictoriaStore) enqueue(batch []model.MetricPayload) {
	//utils.Debug("Enqueue called with %d payloads / %d metrics",		len(batch), totalMetricCount(batch))
	select {
	case v.queue <- batch:
	default:
		utils.Warn("Worker queue full: dropping batch of %d metrics", len(batch))
	}
}

func (v *VictoriaStore) worker() {
	defer v.wg.Done()
	for {
		//utils.Debug(" Worker waiting for batch...")

		select {

		case batch := <-v.queue:
			//utils.Debug(" Worker received batch with %d payloads / %d metrics", len(batch), totalMetricCount(batch))
			v.flush(batch)
		case <-v.ctx.Done():
			utils.Debug("VictoriaStore collector loop exiting")

			return
		}
	}
}

func (v *VictoriaStore) flush(batch []model.MetricPayload) {

	payload := buildPrometheusFormat(batch)

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write([]byte(payload))
	_ = gz.Close()

	//utils.Debug(" Flushing batch of %d metrics", len(batch))

	req, err := http.NewRequest("POST", v.url+"/api/v1/import/prometheus", &buf)
	if err != nil {
		utils.Error("Request build failed: %v", err)
		return
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "text/plain")

	for attempt := 0; attempt < v.batchRetry; attempt++ {
		resp, err := v.client.Do(req)
		if err == nil && resp.StatusCode < 300 {
			//utils.Debug("Batch sent successfully to VictoriaMetrics")
			return
		}
		utils.Warn("Retrying batch write... attempt %d", attempt+1)
		time.Sleep(v.batchInterval)
	}
	utils.Error("Failed to write batch after %d retries", v.batchRetry)
}
