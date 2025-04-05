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

// gosight/agent/internal/bootstrap/metric_store.go
// // Package bootstrap initializes the metric store and metric index for the GoSight agent.
// Package store provides an interface for storing and retrieving metrics.
// It includes an in-memory store and a file-based store for persistence.

package bootstrap

import (
	"fmt"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func InitMetricStore(cfg *config.Config) (store.MetricStore, error) {
	engine := cfg.Storage.Engine
	utils.Info("📦 Initializing metric store engine: %s", engine)

	s, err := store.InitStore(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init metric store: %w", err)
	}

	utils.Info("✅ Metric store [%s] initialized successfully", engine)
	return s, nil
}

func InitMetricIndex() (*store.MetricIndex, error) {

	metricIndex := store.NewMetricIndex()

	return metricIndex, nil
}
