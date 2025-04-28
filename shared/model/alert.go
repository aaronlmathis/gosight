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

import "time"

// AlertRule represents a rule for triggering alerts based on metrics.
type AlertRule struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
	Level      string `json:"level"`      // info, warning, critical
	Expression string `json:"expression"` // e.g. "mem.used_percent > 80 and swap.used_percent > 50"
	Message    string `json:"message"`
	Type       string `json:"type"` // "metric" or "log"

	Match    MatchCriteria `json:"match"`              // metric + label filters
	Actions  []string      `json:"actions"`            // route IDs to trigger
	Cooldown time.Duration `json:"cooldown,omitempty"` // suppress duplicate firing

	EvalInterval    time.Duration `json:"eval_interval"`               // how often to check
	RepeatInterval  time.Duration `json:"repeat_interval,omitempty"`   // e.g. "30m"
	NotifyOnResolve bool          `json:"notify_on_resolve,omitempty"` // true or false
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

// AlertInstance represents the current state of a triggered alert.
type AlertInstance struct {
	ID         string            `json:"id"`
	RuleID     string            `json:"rule_id"`
	EndpointID string            `json:"endpoint_id,omitempty"`
	State      string            `json:"state"`       // "ok", "firing", "resolved", "no_data"
	Previous   string            `json:"previous"`    // previous state
	Scope      string            `json:"scope"`       // "global", "endpoint", "agent", "user", "cloud" etc
	Target     string            `json:"target"`      // e.g. "endpoint_id", "agent_id", "user_id"
	FirstFired time.Time         `json:"first_fired"` // when it first started firing
	LastFired  time.Time         `json:"last_fired"`  // when it last evaluated as firing
	LastOK     time.Time         `json:"last_ok"`     // last time condition returned OK
	LastValue  float64           `json:"last_value"`  // most recent value
	Level      string            `json:"level"`       // from rule (info/warning/critical)
	Message    string            `json:"message"`     // expanded from template
	Labels     map[string]string `json:"labels"`
	ResolvedAt *time.Time        `json:"resolved_at,omitempty"` // when it was resolved
}

type AlertSummary struct {
	RuleID     string    `json:"rule_id"`
	State      string    `json:"state"`       // "firing", "resolved", etc.
	LastChange time.Time `json:"last_change"` // based on LastFired
}
