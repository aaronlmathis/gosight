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

// server/internal/store/registry.go
// Package registry provides a registry for different storage engines.

package metricstore

import (
	"context"
	"fmt"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store/metricindex"
	victoriametricstore "github.com/aaronlmathis/gosight/server/internal/store/metricstore/victoriametrics"

	"github.com/aaronlmathis/gosight/shared/utils"
)

func InitStore(ctx context.Context, cfg *config.Config, metricIndex *metricindex.MetricIndex) (MetricStore, error) {
	utils.Debug("InitMetricStore selected engine: %s", cfg.Storage.Engine)
	switch cfg.Storage.Engine {
	case "victoriametrics":
		utils.Debug("Bootstrapping VictoriaMetrics Store with %d workers", cfg.Storage.Workers)
		s := victoriametricstore.NewVictoriaStore(
			ctx,
			cfg.Storage.URL,
			cfg.Storage.Workers,
			cfg.Storage.QueueSize,
			cfg.Storage.BatchSize,
			cfg.Storage.BatchTimeout,
			cfg.Storage.BatchRetry,
			cfg.Storage.BatchInterval,
			metricIndex,
		)
		utils.Debug("Returning VictoriaStoreMetrics store at: %p", s)
		return s, nil
	default:
		return nil, fmt.Errorf("unsupported storage engine: %s", cfg.Storage.Engine)
	}
}
