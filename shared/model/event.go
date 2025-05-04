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

// // gosight/agent/shared/model/event.go

// Package model defines the data structures used in the GoSight application.
package model

import (
	"time"

	"github.com/google/uuid"
)

// EventEntry represents a single event entry in the event store.
// It includes fields for the event ID, timestamp, level, category, message,
// source, endpoint ID, and any additional metadata.
// The level indicates the severity of the event (e.g., info, warning, critical).
// The category indicates the type of event (e.g., metric, log, system, security).
// The message provides a description of the event.
// The source indicates the origin of the event (e.g., metric name, log source).
// The endpoint ID is used to identify the endpoint associated with the event.
// The meta field is a map of additional tags or labels that can be used to
// provide more context about the event.

type EventEntry struct {
	ID         string            `json:"id"`
	Timestamp  time.Time         `json:"timestamp"`
	Level      string            `json:"level"`    // info, warning, critical
	Type       string            `json:"type"`     // event type (system / alert)
	Category   string            `json:"category"` // metric, log, system, security
	Message    string            `json:"message"`
	Source     string            `json:"source"` // metric name, log source, etc.
	Scope      string            `json:"scope"`  // "endpoint", "system", etc.
	Target     string            `json:"target"` // "host-123", "gosight-core", etc.
	EndpointID string            `json:"endpoint_id"`
	Meta       map[string]string `json:"meta"` // additional tags/labels

}

func GenerateID() string {
	return uuid.NewString()
}

type EventFilter struct {
	Limit      int
	Level      string
	HostID     string
	EndpointID string
	Type       string
	Category   string
	Scope      string
	Target     string
	Source     string
	Contains   string
	Start      *time.Time
	End        *time.Time
	SortOrder  string // "asc" or "desc"
}
