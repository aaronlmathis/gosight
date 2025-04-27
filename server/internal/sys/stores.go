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

// gosight/agent/internal/sys/stores.go
// Package sys provides the system context for the GoSight application.

package sys

import (
	"github.com/aaronlmathis/gosight/server/internal/store/alertstore"
	"github.com/aaronlmathis/gosight/server/internal/store/datastore"
	"github.com/aaronlmathis/gosight/server/internal/store/eventstore"
	"github.com/aaronlmathis/gosight/server/internal/store/logstore"
	"github.com/aaronlmathis/gosight/server/internal/store/metricstore"
	"github.com/aaronlmathis/gosight/server/internal/store/routestore"
	"github.com/aaronlmathis/gosight/server/internal/store/rulestore"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
)

// StoreModule contains all persistent or semi-persistent storage components.
type StoreModule struct {
	Metrics metricstore.MetricStore // Time-series metrics (e.g., VictoriaMetrics)
	Logs    logstore.LogStore       // Structured logs (e.g., journald, /var/log/secure)
	Users   userstore.UserStore     // User auth, roles, permissions
	Data    datastore.DataStore     // Hosts, endpoints, agents, etc.
	Events  eventstore.EventStore   // Event logs, audit, alert events
	Rules   rulestore.RuleStore     // Alert rules defined by users
	Actions *routestore.RouteStore  // Routes loaded from alert_routes.yaml
	Alerts  alertstore.AlertStore   // Alert instances (active, resolved, history)
}

// NewStoreModule creates a new StoreModule with the provided components.

func NewStoreModule(
	metrics metricstore.MetricStore,
	logs logstore.LogStore,
	users userstore.UserStore,
	data datastore.DataStore,
	events eventstore.EventStore,
	rules rulestore.RuleStore,
	actions *routestore.RouteStore,
	alerts alertstore.AlertStore,
) *StoreModule {
	return &StoreModule{
		Metrics: metrics,
		Logs:    logs,
		Users:   users,
		Data:    data,
		Events:  events,
		Rules:   rules,
		Actions: actions,
		Alerts:  alerts,
	}
}
