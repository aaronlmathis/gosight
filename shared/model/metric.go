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

package model

import "time"

type StatisticValues struct {
	Minimum     float64 `json:"min"`
	Maximum     float64 `json:"max"`
	SampleCount int     `json:"count"`
	Sum         float64 `json:"sum"`
}

type Point struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

type Metric struct {
	Namespace         string            `json:"namespace,omitempty"`
	SubNamespace      string            `json:"subnamespace,omitempty"`
	Name              string            `json:"name"`
	Timestamp         time.Time         `json:"timestamp,omitempty"`
	Value             float64           `json:"value,omitempty"`
	StatisticValues   *StatisticValues  `json:"stats,omitempty"`
	Unit              string            `json:"unit,omitempty"`
	Dimensions        map[string]string `json:"dimensions,omitempty"`
	StorageResolution int               `json:"resolution,omitempty"`
	Type              string            `json:"type,omitempty"`
}

type MetricPayload struct {
	AgentID    string    `json:"agent_id"`
	HostID     string    `json:"host_id"`
	Hostname   string    `json:"hostname"`
	EndpointID string    `json:"endpoint_id"`
	Metrics    []Metric  `json:"metrics"`
	Meta       *Meta     `json:"meta,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

type MetricRow struct {
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags"`
	Timestamp int64             `json:"timestamp"` // Unix ms
}

type MetricPoint struct {
	Timestamp int64   `json:"timestamp"` // Unix ms
	Value     float64 `json:"value"`     // Metric value
}

type MetricSelector struct {
	Name         string
	Namespace    string
	SubNamespace string
	Instant      bool // true = stat card, false = chart or long term
}
