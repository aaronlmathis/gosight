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

// ActionRouteSet represents a set of action routes for alerting.
// It contains a list of ActionRoute objects, each defining a specific route
// for handling alerts based on matching criteria.
type ActionRouteSet struct {
	Routes []ActionRoute `yaml:"routes" json:"routes"`
}

// ActionRoute defines a single action route for alerting.
// It includes a unique ID, a match filter to determine when the route should be triggered,
// and a list of actions to be executed when the route is matched.
type ActionRoute struct {
	ID      string       `yaml:"id" json:"id"`
	Match   MatchFilter  `yaml:"match" json:"match"`
	Actions []ActionSpec `yaml:"actions" json:"actions"`
}

// MatchFilter defines the criteria for matching alerts.
type MatchFilter struct {
	Level  string            `yaml:"level,omitempty" json:"level,omitempty"`     // warning, critical, etc
	RuleID string            `yaml:"rule_id,omitempty" json:"rule_id,omitempty"` // specific rule
	Tags   map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`       // meta.Tags match
}

// ActionSpec defines the specifications for an action to be taken when an alert is triggered.
// It includes the type of action (e.g., webhook, email, script), the URL for webhooks,
// headers for the request, and the command to be executed for scripts.
type ActionSpec struct {
	Type    string            `yaml:"type" json:"type"` // webhook, email, script
	URL     string            `yaml:"url,omitempty"`    // for webhook
	Headers map[string]string `yaml:"headers,omitempty"`

	Command string   `yaml:"command,omitempty"` // for script
	Args    []string `yaml:"args,omitempty"`    // optional args
}
