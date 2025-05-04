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
	ID          string        `json:"id" yaml:"id"`
	Name        string        `json:"name" yaml:"name"`
	Description string        `json:"description,omitempty" yaml:"description,omitempty"`
	Message     string        `json:"message" yaml:"message"` // message template for alert
	Level       string        `json:"level" yaml:"level"`     // info, warning, critical
	Enabled     bool          `json:"enabled" yaml:"enabled"`
	Type        string        `json:"type" yaml:"type"` // metric, log, event, composite
	Match       MatchCriteria `json:"match" yaml:"match"`
	Scope       Scope         `json:"scope" yaml:"scope"`
	Expression  Expression    `json:"expression" yaml:"expression"`
	Actions     []string      `json:"actions" yaml:"actions"`
	Options     Options       `json:"options" yaml:"options"`
}

type Scope struct {
	Namespace    string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	SubNamespace string `json:"subnamespace,omitempty" yaml:"subnamespace,omitempty"`
	Metric       string `json:"metric,omitempty" yaml:"metric,omitempty"`
}

type Expression struct {
	Operator string      `json:"operator" yaml:"operator"`                     // >, <, =, !=, contains, regex
	Value    interface{} `json:"value" yaml:"value"`                           // number or string
	Datatype string      `json:"datatype,omitempty" yaml:"datatype,omitempty"` // percent, numeric, status
}

type Options struct {
	Cooldown        string `json:"cooldown,omitempty" yaml:"cooldown,omitempty"`
	EvalInterval    string `json:"eval_interval,omitempty" yaml:"eval_interval,omitempty"`
	RepeatInterval  string `json:"repeat_interval,omitempty" yaml:"repeat_interval,omitempty"`
	NotifyOnResolve bool   `json:"notify_on_resolve,omitempty" yaml:"notify_on_resolve,omitempty"`
}

type MatchCriteria struct {
	EndpointIDs []string          `json:"endpoint_ids,omitempty" yaml:"endpoint_ids,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Category    string            `json:"category,omitempty" yaml:"category,omitempty"`
	Source      string            `json:"source,omitempty" yaml:"source,omitempty"`
	Scope       string            `json:"scope,omitempty" yaml:"scope,omitempty"`
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

type AlertQuery struct {
	RuleID string
	State  string
	Level  string
	Target string
	Scope  string
	Sort   string // e.g. "first_fired desc"
	Limit  int
	Offset int
	Order  string
	Tags   map[string]string
}
