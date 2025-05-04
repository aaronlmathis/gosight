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
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
)

type VictoriaStore struct {
	url    string
	client *http.Client
}

// NewVictoriaStore
func NewVictoriaStore(url string) (*VictoriaStore, error) {
	return &VictoriaStore{
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (v *VictoriaStore) Write(batch []model.MetricPayload) error {
	if len(batch) == 0 {
		return nil
	}

	payload := buildPrometheusFormat(batch)

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write([]byte(payload))
	_ = gz.Close()

	req, err := http.NewRequest("POST", v.url+"/api/v1/import/prometheus", &buf)
	if err != nil {
		return fmt.Errorf("VM request build failed: %w", err)
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "text/plain")

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("VM request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("VM response error: %s", string(body))
	}

	return nil
}

func (v *VictoriaStore) Close() error {
	// No resources to release
	// May Need in future
	return nil
}
