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

// server/internal/bootstrap/caches.go

package bootstrap

import (
	"context"
	"fmt"

	"github.com/aaronlmathis/gosight/server/internal/cache"
	"github.com/aaronlmathis/gosight/server/internal/store/datastore"
)

// InitCaches initializes caches for the system context.
func InitCaches(ctx context.Context, dataStore datastore.DataStore) (*cache.Cache, error) {
	caches := &cache.Cache{}

	// Initialize Metric Cache
	metricCache := cache.NewMetricCache()
	caches.Metrics = metricCache

	// Initialize tag cache
	tagCache := cache.NewTagCache()

	// Warm fill tag cache
	tags, err := dataStore.GetAllTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load tags for cache: %w", err)
	}

	tagCache.LoadFromStore(tags)

	caches.Tags = tagCache

	// Initialize Process Cache
	processCache := cache.NewProcessCache()
	caches.Processes = processCache

	// Log Cache

	logCache := cache.NewLogCache()
	caches.Logs = logCache

	return caches, nil
}
