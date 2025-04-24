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

// server/internal/bootstrap/agent.go
// Init agent tracking from in-memory

package bootstrap

import (
	"context"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/events"
	"github.com/aaronlmathis/gosight/server/internal/store/agenttracker"
	"github.com/aaronlmathis/gosight/server/internal/store/datastore"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// InitAgentTracker initializes the agent tracker for the GoSight agent.
// The agent tracker is responsible for tracking the state of agents and
// their associated metrics and logs.
func InitAgentTracker(ctx context.Context, env string, dataStore datastore.DataStore, emitter *events.Emitter) (*agenttracker.AgentTracker, error) {
	tracker := agenttracker.NewAgentTracker(ctx, emitter, dataStore)

	//  Start sync loop to periodically push in-memory AgentTracker data into persistant store.
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				utils.Info("Agent tracker sync loop shutting down")
				return
			case <-ticker.C:
				//utils.Debug("Syncing agent tracker to DB...")
				tracker.SyncToStore(ctx, dataStore)
				//utils.Debug("Agent tracker sync complete")
			}
		}
	}()
	// Start a loop to check agent statuses and emit events
	// This loop will run every 30 seconds and check the status of all agents
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				utils.Info("Agent tracker status check loop shutting down")
				return
			case <-ticker.C:
				tracker.CheckAgentStatusesAndEmitEvents()
			}
		}
	}()
	return tracker, nil
}
