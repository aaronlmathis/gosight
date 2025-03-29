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
along with LeetScraper. If not, see https://www.gnu.org/licenses/.
*/

package collector

import (
	"sync"
)

type MetricStore struct {
	mu      sync.RWMutex
	metrics map[string]float64
}

func NewMetricStore() *MetricStore {
	return &MetricStore{
		metrics: make(map[string]float64),
	}
}

// Update the store with new metrics
func (s *MetricStore) Update(results map[string]float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range results {
		s.metrics[k] = v
	}
}

// Get snapshot of all current metrics
func (s *MetricStore) Snapshot() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copy := make(map[string]float64)
	for k, v := range s.metrics {
		copy[k] = v
	}
	return copy
}
