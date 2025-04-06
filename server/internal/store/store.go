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

// gosight/server/internal/store/store.go
// Package store provides an interface for writing metrics to different storage engines.
// It defines the MetricStore interface, which includes methods for writing metrics
// and closing the store connection. This allows for flexibility in choosing the
// underlying storage engine, such as VictoriaMetrics or others in the future.

package store

import (
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
)

type MetricStore interface {
	Write(metrics []model.MetricPayload) error
	Close() error

	QueryInstant(metric string, filters map[string]string) ([]model.MetricRow, error)
	QueryRange(metric string, start, end time.Time, filters map[string]string) ([]model.Point, error)
}
