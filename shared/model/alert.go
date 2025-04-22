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

// shared/model/alert.go

package model

// AlertRule represents a rule for triggering alerts based on metrics.
type AlertRule struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Enabled bool             `json:"enabled"`
	Level   string           `json:"level"` // info, warning, critical
	Message string           `json:"message"`
	Match   MatchCriteria    `json:"match"`
	Trigger TriggerCondition `json:"trigger"`
}

// MetricSelector defines the metric to be monitored and any labels to match.
type MatchCriteria struct {
	Namespace    string            `json:"namespace,omitempty"`
	SubNamespace string            `json:"subnamespace,omitempty"`
	Metric       string            `json:"metric,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`        // e.g. container_name=webapp
	EndpointIDs  []string          `json:"endpoint_ids,omitempty"`  // specific targets
	TagSelectors map[string]string `json:"tag_selectors,omitempty"` // match Meta.Tags
}

// TriggerCondition defines the condition under which an alert is triggered.
type TriggerCondition struct {
	Operator  string  `json:"operator"` // "gt", "lt", "eq"
	Threshold float64 `json:"threshold"`
	Duration  string  `json:"duration"` // e.g., "5m"
}

// AlertInstance represents the current state of a triggered alert.
type AlertInstance struct {
	RuleID     string            `json:"rule_id"`
	EndpointID string            `json:"endpoint_id"`
	State      string            `json:"state"`       // "firing", "resolved"
	FirstFired string            `json:"first_fired"` // RFC3339
	LastFired  string            `json:"last_fired"`  // RFC3339
	LastValue  float64           `json:"last_value"`  // latest metric value
	Labels     map[string]string `json:"labels"`      // optional: container, interface, etc.
	Message    string            `json:"message"`     // from rule
	Level      string            `json:"level"`       // info/warning/critical
}
