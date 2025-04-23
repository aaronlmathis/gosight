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

// gosight/agent/internal/sys/sys.go
// Package sys provides the system context for the GoSight application.
package sys

import (
	"context"

	"github.com/aaronlmathis/gosight/server/internal/alerts"
	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/dispatcher"
	"github.com/aaronlmathis/gosight/server/internal/events"
	"github.com/aaronlmathis/gosight/server/internal/rules"
	"github.com/aaronlmathis/gosight/server/internal/store/agenttracker"
	"github.com/aaronlmathis/gosight/server/internal/store/alertstore"
	"github.com/aaronlmathis/gosight/server/internal/store/datastore"
	"github.com/aaronlmathis/gosight/server/internal/store/eventstore"
	"github.com/aaronlmathis/gosight/server/internal/store/logstore"
	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/aaronlmathis/gosight/server/internal/store/metricindex"
	"github.com/aaronlmathis/gosight/server/internal/store/metricstore"
	"github.com/aaronlmathis/gosight/server/internal/store/routestore"
	"github.com/aaronlmathis/gosight/server/internal/store/rulestore"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/server/internal/websocket"
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

// TelemetryModule encapsulates telemetry-related state and processing.
type TelemetryModule struct {
	Index      *metricindex.MetricIndex // Metric name/dimension catalog
	Meta       *metastore.MetaTracker   // Tracks source metadata (labels, tags, endpoint info)
	Evaluator  *rules.Evaluator         // Rule evaluator (metrics → match?)
	Alerts     *alerts.Manager          // Tracks alert state per rule/endpoint
	Emitter    *events.Emitter          // Emits events (alerts, system actions)
	Dispatcher *dispatcher.Dispatcher   // Routes alert events to actions
}

// SystemContext is passed to all subsystems, providing full access to config, state, and interfaces.
type SystemContext struct {
	Ctx    context.Context
	Cfg    *config.Config
	Agents *agenttracker.AgentTracker // Tracks agent state, uptime, heartbeat
	Web    *websocket.Hub             // WebSocket hub for live streaming to UI
	Auth   map[string]gosightauth.AuthProvider
	Stores *StoreModule
	Tele   *TelemetryModule
}
