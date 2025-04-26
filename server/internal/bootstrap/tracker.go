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
	"github.com/aaronlmathis/gosight/server/internal/store/datastore"
	"github.com/aaronlmathis/gosight/server/internal/tracker"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// InitEndpointTracker initializes the unified endpoint tracker.
// Tracks both agents and containers, and emits lifecycle events.
func InitTracker(ctx context.Context, dataStore datastore.DataStore, emitter *events.Emitter) *tracker.EndpointTracker {
	t := tracker.NewEndpointTracker(ctx, emitter, dataStore)

	// Sync to DB every 60s
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				utils.Info("EndpointTracker sync loop shutting down")
				return
			case <-ticker.C:
				t.SyncToStore(ctx, dataStore)
			}
		}
	}()

	// Emit lifecycle events every 30s
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				utils.Info("EndpointTracker status check loop shutting down")
				return
			case <-ticker.C:
				t.CheckAgentStatusesAndEmitEvents()
				t.CheckContainerStatusesAndEmit()
			}
		}
	}()

	return t
}
