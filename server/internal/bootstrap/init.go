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

// gosight/agent/internal/bootstrap/init.go

package bootstrap

import (
	"context"
	"fmt"

	"github.com/aaronlmathis/gosight/server/internal/alerts"
	"github.com/aaronlmathis/gosight/server/internal/dispatcher"
	"github.com/aaronlmathis/gosight/server/internal/events"
	"github.com/aaronlmathis/gosight/server/internal/rules"
	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/aaronlmathis/gosight/server/internal/syncmanager"
	"github.com/aaronlmathis/gosight/server/internal/sys"
	"github.com/aaronlmathis/gosight/server/internal/tracker"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// initGoSight initializes the GoSight server by setting up various components
// such as logging, metric index, log store, metric store, data store,
// agent tracker, meta tracker, websocket hub, user store, event store,
// rule store, emitter/alert manager, evaluator, and authentication providers.

// It loads these components into a SystemContext and returns an error if any
// of the components fail to initialize.

func InitGoSight(ctx context.Context) (*sys.SystemContext, error) {

	// Initialize the GoSight server
	fmt.Println("Initializing GoSight server...")

	// Load the configuration
	// Bootstrap config loading (flags -> env -> file)
	cfg := LoadServerConfig()

	fmt.Printf("About to init logger with level = %s\n", cfg.Logs.LogLevel)

	// Initialize logging
	if err := utils.InitLogger(cfg.Logs.AppLogFile, cfg.Logs.ErrorLogFile, cfg.Logs.AccessLogFile, cfg.Logs.DebugLogFile, cfg.Logs.LogLevel); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize metric index
	metricIndex, err := InitMetricIndex()
	utils.Must("Metric index", err)

	// Initialize data store
	dataStore, err := InitDataStore(cfg)
	utils.Must("Data store", err)

	// Initialize meta tracker
	metaTracker := metastore.NewMetaTracker()

	// Initialize the websocket hub
	wsHub := InitWebSocketHub(ctx, metaTracker)

	// Initialize event store
	eventStore, err := InitEventStore(cfg)
	utils.Must("Event store", err)

	// Initialize alert store
	alertStore, err := InitAlertStore(cfg)
	utils.Must("Alert store", err)

	// Initialize rule store
	ruleStore, err := InitRuleStore(cfg)
	utils.Must("Rule store", err)

	// Initialize action store
	actionStore, err := InitRouteStore(cfg)
	utils.Must("Action store", err)

	// Initialize emitter
	emitter := events.NewEmitter(eventStore, wsHub.Events)

	// Initialize dispatcher
	dispatcher := dispatcher.NewDispatcher(actionStore.BuildMap())

	// Initialize alert manager
	alertMgr := alerts.NewManager(emitter, dispatcher, alertStore, wsHub.Alerts)

	// Initialize the evaluator
	evaluator := rules.NewEvaluator(ruleStore, alertMgr)

	// Initialize user store
	userStore, err := InitUserStore(cfg)
	utils.Must("User store", err)

	// Initialize auth
	authProviders, err := InitAuth(cfg, userStore)
	utils.Must("Auth providers", err)

	// Initialize agent tracker
	tracker := tracker.NewEndpointTracker(ctx, emitter, dataStore)
	utils.Must("Agent tracker", err)

	// Initialize cache
	caches, err := InitCaches(ctx, dataStore)
	utils.Must("Caches", err)

	// Init metric store
	metricStore, err := InitMetricStore(ctx, cfg, caches.Metrics)
	utils.Must("Metric store", err)

	// Initialize log store
	logStore, err := InitLogStore(ctx, cfg, caches.Logs)
	utils.Must("Log store", err)

	// Initialize SyncManager (synchronization of caches with datastore)
	syncManager := syncmanager.NewSyncManager(ctx, caches, dataStore, tracker, cfg.Server.SyncInterval)

	// Build stores
	stores := sys.NewStoreModule(
		metricStore,
		logStore,
		userStore,
		dataStore,
		eventStore,
		ruleStore,
		actionStore,
		alertStore,
	)
	buffers := InitBufferEngine(ctx, &cfg.BufferEngine, stores)
	// Build telemetry
	telemetry := sys.NewTelemetryModule(
		metricIndex,
		metaTracker,
		evaluator,
		alertMgr,
		emitter,
		dispatcher,
	)

	// Initialize the system context
	// The system context holds all the components of the GoSight server
	// and provides a way to access them throughout the application.
	sys := sys.NewSystemContext(
		ctx,
		cfg,
		tracker,
		wsHub,
		authProviders,
		stores,
		telemetry,
		caches,
		buffers,
		syncManager,
	)

	return sys, nil

}
