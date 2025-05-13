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

// server/internal/store/logstore/victoriametrics/victoriametrics.go

package victorialogstore

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/cache"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type VictoriaLogStore struct {
	url      string
	client   *http.Client
	logsPath string
	cache    cache.LogCache
}

// NewVictoriaLogStore
func NewVictoriaLogStore(url string, dir string, logCache cache.LogCache) (*VictoriaLogStore, error) {
	return &VictoriaLogStore{
		url:      url,
		client:   &http.Client{Timeout: 10 * time.Second},
		logsPath: dir,
		cache:    logCache,
	}, nil
}

func (v *VictoriaLogStore) Name() string {
	return "VictoriaMetrics LogStore"
}

func (v *VictoriaLogStore) Write(batch []model.LogPayload) error {
	if len(batch) == 0 {
		return nil
	}

	// Wrap all logs once — generates log_id and preserves Meta
	wrapped := wrapLogs(batch)

	// Store in cache
	v.cache.Add(wrapped)

	// Write full logs to compressed JSON (one entry per line)
	if err := v.writeCompressedWrappedLogs(wrapped); err != nil {
		utils.Warn("Failed to write logs to JSON: %v", err)
	}

	// Generate Prometheus format using consistent log_id and labels
	payload := buildPrometheusFormatFromWrapped(wrapped)

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write([]byte(payload))
	_ = gz.Close()

	req, err := http.NewRequest("POST", v.url+"/api/v1/import/prometheus", &buf)
	if err != nil {
		utils.Debug("Failed to create VM request: %v", err)
		return fmt.Errorf("VM request build failed: %w", err)
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "text/plain")

	resp, err := v.client.Do(req)
	if err != nil {
		utils.Debug("Failed to create VM request: %v", err)
		return fmt.Errorf("VM request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("VM response error: %s", string(body))
	}
	utils.Debug("VictoriaMetrics response: %s", resp.Status)
	return nil
}

func (v *VictoriaLogStore) Close() error {
	// No resources to release
	// May Need in future
	return nil
}
