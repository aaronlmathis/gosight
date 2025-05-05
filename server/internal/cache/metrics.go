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

package cache

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
)

// MetricCache is an in-memory cache for metrics as well as their meta data
type MetricCache interface {
	Add(payload *model.MetricPayload)
	GetNamespaces() []string                                // Get all known namespaces
	GetSubNamespaces(nameSpace string) []string             // Get all subnamespaces for a namespace
	GetMetricNames(nameSpace, subNamespace string) []string // Get all metric names for Namespace+Subnamespace

	GetAllMetricNames() []string                    // Get all known metric names
	GetAvailableDimensions() map[string][]string    // Get all available dimensions (known)
	GetMetricDimensions(metricName string) []string // Get all dimension keys known for a metric

	GetAllTagKeys() []string                   // Get all known tag keys
	GetAllTagValuesForKey(key string) []string // Get all known values for tag key

	// Label = Metric dimension + Tags (prometheus labels)
	GetAllKnownLabelValues(label, contains string) []string  // Get all known values for a given label key (dimensions + tags) (optionally filtered)
	GetLabelValues(label, contains string) []string          // Get all label values for a known label (optionally filtered)
	GetMetricsWithLabels(filters map[string]string) []string // Get all metric names that match a given label filter

	Prune()
}

type StringSet map[string]struct{}

func addToSet(m map[string]StringSet, key, val string) {
	if m[key] == nil {
		m[key] = make(StringSet)
	}
	m[key][val] = struct{}{}
}

type MetricEntry struct {
	Namespace string
	SubNS     string
	Name      string
	Unit      string
	Type      string

	Dimensions map[string]StringSet
	Labels     map[string]StringSet
	Emitters   map[string]struct{} // endpoint IDs
}
type metricCache struct {
	ctx context.Context
	mu  sync.RWMutex

	Namespaces    map[string]struct{}
	SubNamespaces map[string]map[string]struct{} // ns → subns
	MetricEntries map[string]*MetricEntry        // fullName → entry
	LabelValues   map[string]StringSet           // label key → all values
	EndpointMeta  map[string]*model.Meta         // endpoint_id → latest Meta
	LastSeen      map[string]int64
}

func NewMetricCache(ctx context.Context) MetricCache {
	return &metricCache{
		ctx:           ctx,
		Namespaces:    make(map[string]struct{}),
		SubNamespaces: make(map[string]map[string]struct{}),
		MetricEntries: make(map[string]*MetricEntry),
		LabelValues:   make(map[string]StringSet),
		EndpointMeta:  make(map[string]*model.Meta),
		LastSeen:      make(map[string]int64),
	}
}

func (c *metricCache) Add(payload *model.MetricPayload) {
	c.mu.Lock()
	defer c.mu.Unlock()

	eid := payload.EndpointID
	if payload.Meta != nil {
		c.EndpointMeta[eid] = payload.Meta
	}
	c.LastSeen[eid] = time.Now().Unix()
	for _, m := range payload.Metrics {
		fullName := m.Namespace + "." + m.SubNamespace + "." + m.Name

		if _, ok := c.Namespaces[m.Namespace]; !ok {
			c.Namespaces[m.Namespace] = struct{}{}
		}
		if _, ok := c.SubNamespaces[m.Namespace]; !ok {
			c.SubNamespaces[m.Namespace] = make(map[string]struct{})
		}
		c.SubNamespaces[m.Namespace][m.SubNamespace] = struct{}{}

		entry := c.MetricEntries[fullName]
		if entry == nil {
			entry = &MetricEntry{
				Namespace:  m.Namespace,
				SubNS:      m.SubNamespace,
				Name:       m.Name,
				Unit:       m.Unit,
				Type:       m.Type,
				Dimensions: make(map[string]StringSet),
				Labels:     make(map[string]StringSet),
				Emitters:   make(map[string]struct{}),
			}
			c.MetricEntries[fullName] = entry
		}

		entry.Emitters[eid] = struct{}{}

		// Index dimensions
		for k, v := range m.Dimensions {
			addToSet(entry.Dimensions, k, v)
			addToSet(entry.Labels, k, v)
			addToSet(c.LabelValues, k, v)
		}

		// Index tags
		if payload.Meta != nil {
			for k, v := range payload.Meta.Tags {
				addToSet(entry.Labels, k, v)
				addToSet(c.LabelValues, k, v)
			}
		}
	}
}

func (c *metricCache) GetNamespaces() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var out []string
	for ns := range c.Namespaces {
		out = append(out, ns)
	}
	return out
}

func (c *metricCache) GetMetricNames(namespace, subNamespace string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var names []string
	for _, entry := range c.MetricEntries {
		if entry.Namespace == namespace && entry.SubNS == subNamespace {
			names = append(names, entry.Name)
		}
	}
	return names
}

func (c *metricCache) GetSubNamespaces(namespace string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	subnsMap, ok := c.SubNamespaces[namespace]
	if !ok {
		return nil
	}

	var out []string
	for sub := range subnsMap {
		out = append(out, sub)
	}
	return out
}

func (c *metricCache) GetAllMetricNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var names []string
	for fullName := range c.MetricEntries {
		names = append(names, fullName)
	}
	return names
}

func (c *metricCache) GetAvailableDimensions() map[string][]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make(map[string][]string)
	for key, valueSet := range c.LabelValues {
		// Only include keys that are used as dimensions (not tags only)
		isDimension := false
		for _, entry := range c.MetricEntries {
			if _, ok := entry.Dimensions[key]; ok {
				isDimension = true
				break
			}
		}
		if isDimension {
			var values []string
			for v := range valueSet {
				values = append(values, v)
			}
			out[key] = values
		}
	}
	return out
}

func (c *metricCache) GetMetricDimensions(metricName string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.MetricEntries[metricName]
	if !ok {
		return nil
	}

	var keys []string
	for k := range entry.Dimensions {
		keys = append(keys, k)
	}
	return keys
}

func (c *metricCache) GetAllTagKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Collect tag keys directly from EndpointMeta
	tagKeySet := make(map[string]struct{})

	for _, meta := range c.EndpointMeta {
		for k := range meta.Tags {
			tagKeySet[k] = struct{}{}
		}
	}

	var keys []string
	for k := range tagKeySet {
		keys = append(keys, k)
	}
	return keys
}
func (c *metricCache) GetAllTagValuesForKey(key string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	valueSet := make(map[string]struct{})

	for _, meta := range c.EndpointMeta {
		if val, ok := meta.Tags[key]; ok {
			valueSet[val] = struct{}{}
		}
	}

	var values []string
	for v := range valueSet {
		values = append(values, v)
	}
	return values
}

func (c *metricCache) GetAllKnownLabelValues(label, contains string) []string {
	return c.GetLabelValues(label, contains)
}

func (c *metricCache) GetLabelValues(label, contains string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	values, ok := c.LabelValues[label]
	if !ok {
		return nil
	}

	var out []string
	for v := range values {
		if contains == "" || contains == v || containsMatch(v, contains) {
			out = append(out, v)
		}
	}
	return out
}

func (c *metricCache) GetMetricsWithLabels(filters map[string]string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var matched []string

	for fullName, entry := range c.MetricEntries {
		match := true

		for key, expected := range filters {
			values, exists := entry.Labels[key]
			if !exists {
				match = false
				break
			}
			if _, valueExists := values[expected]; !valueExists {
				match = false
				break
			}
		}

		if match {
			matched = append(matched, fullName)
		}
	}

	return matched
}
func (c *metricCache) Prune() {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := time.Now().Add(-1440 * time.Minute).Unix() // e.g. 10 minutes (one day 1440 minutes)

	// Identify stale endpoints
	stale := make(map[string]struct{})
	for eid, ts := range c.LastSeen {
		if ts < cutoff {
			stale[eid] = struct{}{}
			delete(c.EndpointMeta, eid)
			delete(c.LastSeen, eid)
		}
	}

	// Remove emitters from metrics
	for fullName, entry := range c.MetricEntries {
		for eid := range stale {
			delete(entry.Emitters, eid)
		}
		// Optional: remove metrics with no active emitters
		if len(entry.Emitters) == 0 {
			delete(c.MetricEntries, fullName)
		}
	}
}
func containsMatch(value, substr string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(substr))
}
