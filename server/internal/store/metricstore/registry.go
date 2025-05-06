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

	"github.com/aaronlmathis/gosight/server/internal/cache"
	"github.com/aaronlmathis/gosight/server/internal/config"
	victoriametricstore "github.com/aaronlmathis/gosight/server/internal/store/metricstore/victoriametrics"

	"github.com/aaronlmathis/gosight/shared/utils"
)

func InitMetricStore(ctx context.Context, cfg *config.Config, metricCache cache.MetricCache) (MetricStore, error) {

	switch cfg.MetricStore.Engine {
	case "victoriametrics":
		s, err := victoriametricstore.NewVictoriaStore(cfg.MetricStore.URL, metricCache)
		if err != nil {
			return nil, err
		}
		utils.Debug("Returning VictoriaStoreMetrics store at: %p", s)
		return s, nil
	default:
		return nil, fmt.Errorf("unsupported storage engine: %s", cfg.MetricStore.Engine)
	}
}
