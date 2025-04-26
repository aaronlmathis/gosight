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

// gosight/agent/internal/store/metric_index.go
// Package store provides an interface for storing and retrieving metrics.
// It includes an in-memory store and a file-based store for persistence.

package metricindex

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/aaronlmathis/gosight/shared/utils"
)

type MetricIndex struct {
	mu               sync.RWMutex
	Namespaces       map[string]struct{}
	SubNamespaces    map[string]map[string]struct{}            // namespace → subnamespace
	MetricNames      map[string]map[string]map[string]struct{} // ns → sub → metric names
	Dimensions       map[string]map[string]struct{}            // dim key → value set
	MetricDimensions map[string]map[string]string              // metricFullName → dim key → value

}

func NewMetricIndex() *MetricIndex {
	return &MetricIndex{
		Namespaces:       make(map[string]struct{}),
		SubNamespaces:    make(map[string]map[string]struct{}),
		MetricNames:      make(map[string]map[string]map[string]struct{}),
		Dimensions:       make(map[string]map[string]struct{}),
		MetricDimensions: make(map[string]map[string]string), // metricFullName → dim key → value

	}
}

func (idx *MetricIndex) Add(namespace, sub, name string, dims map[string]string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	normNamespace := strings.ToLower(namespace)
	normSub := strings.ToLower(sub)
	normName := strings.ToLower(name)

	idx.Namespaces[normNamespace] = struct{}{}

	if _, ok := idx.SubNamespaces[normNamespace]; !ok {
		idx.SubNamespaces[normNamespace] = make(map[string]struct{})
	}
	idx.SubNamespaces[normNamespace][normSub] = struct{}{}

	if _, ok := idx.MetricNames[normNamespace]; !ok {
		idx.MetricNames[normNamespace] = make(map[string]map[string]struct{})
	}
	if _, ok := idx.MetricNames[normNamespace][normSub]; !ok {
		idx.MetricNames[normNamespace][normSub] = make(map[string]struct{})
	}
	fullName := fmt.Sprintf("%s.%s.%s", normNamespace, normSub, normName)
	idx.MetricNames[normNamespace][normSub][fullName] = struct{}{}

	for k, v := range dims {
		if _, ok := idx.Dimensions[k]; !ok {
			idx.Dimensions[k] = make(map[string]struct{})
		}
		idx.Dimensions[k][v] = struct{}{}
	}

	if idx.MetricDimensions == nil {
		idx.MetricDimensions = make(map[string]map[string]string)
	}
	idx.MetricDimensions[fullName] = dims
	//	utils.Debug("🔢 Indexed metric %s with dimensions: %+v", fullName, dims)
}

func (idx *MetricIndex) GetNamespaces() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	out := make([]string, 0, len(idx.Namespaces))
	for k := range idx.Namespaces {
		out = append(out, k)
	}
	return out
}

func (idx *MetricIndex) GetSubNamespaces(ns string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	sub, ok := idx.SubNamespaces[ns]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(sub))
	for s := range sub {
		out = append(out, s)
	}
	return out
}

func (idx *MetricIndex) GetAllMetricNames() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	seen := make(map[string]struct{})
	for _, subMap := range idx.MetricNames {
		for _, nameSet := range subMap {
			for n := range nameSet {
				seen[n] = struct{}{}
			}
		}
	}
	var out []string
	for n := range seen {
		out = append(out, n)
	}
	return out
}

func (idx *MetricIndex) GetMetricNames(ns, sub string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	names, ok := idx.MetricNames[ns][sub]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(names))
	for n := range names {
		out = append(out, n)
	}
	return out
}

func (idx *MetricIndex) GetDimensions() map[string][]string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	out := make(map[string][]string)
	for k, vs := range idx.Dimensions {
		for v := range vs {
			out[k] = append(out[k], v)
		}
	}
	return out
}

func (idx *MetricIndex) GetDimensionsForMetric(fullMetric string) ([]string, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	dims, ok := idx.MetricDimensions[fullMetric]
	if !ok {
		return nil, fmt.Errorf("metric %s not found in MetricDimensions", fullMetric)
	}

	var keys []string
	for k := range dims {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

// FilterMetricNames returns all metric names that match given label filters
func (idx *MetricIndex) FilterMetricNames(filters map[string]string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	matched := make(map[string]struct{})

	for _, subMap := range idx.MetricNames {
		for _, nameSet := range subMap {
			for name := range nameSet {
				// name is already fullName
				if idx.matchesAnyDimension(name, filters) {
					matched[name] = struct{}{}
				}
			}
		}
	}

	var result []string
	for name := range matched {
		result = append(result, name)
	}
	return result
}

func (idx *MetricIndex) matchesAnyDimension(name string, filters map[string]string) bool {

	dims, ok := idx.MetricDimensions[name]
	if !ok {
		utils.Debug("❌ No dimensions found for %s", name)
		return false
	}
	//utils.Debug("🔍 Comparing dims=%v with filters=%v", dims, filters)

	for k, v := range filters {
		if dims[k] != v {
			return false
		}
	}
	return true
}
