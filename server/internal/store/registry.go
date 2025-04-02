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

package store

import (
	"fmt"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func InitStore(cfg config.StorageConfig) (MetricStore, error) {
	utils.Debug("🧠 InitMetricStore selected engine: %s", cfg.Engine)
	switch cfg.Engine {
	case "victoriametrics":
		utils.Debug("📦 Bootstrapping VictoriaStore with %d workers", cfg.Workers)
		s := NewVictoriaStore(
			cfg.URL,
			cfg.Workers,
			cfg.QueueSize,
			cfg.BatchSize,
			cfg.BatchTimeout,
			cfg.BatchRetry,
			cfg.BatchInterval,
		)
		utils.Debug("✅ Returning VictoriaStore at: %p", s)
		return s, nil
	default:
		return nil, fmt.Errorf("unsupported storage engine: %s", cfg.Engine)
	}
}
