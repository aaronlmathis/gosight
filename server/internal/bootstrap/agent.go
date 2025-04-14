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

	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/server/internal/store/datastore"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func InitAgentTracker(ctx context.Context, env string, dataStore datastore.DataStore) (*store.AgentTracker, error) {
	tracker := store.NewAgentTracker()

	//  Start sync loop to periodically push in-memory AgentTracker data into persistant store.
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				utils.Info("Agent tracker sync loop shutting down")
				return
			case <-ticker.C:
				utils.Debug("Syncing agent tracker to DB...")
				tracker.SyncToStore(ctx, dataStore)
				utils.Debug("Agent tracker sync complete")
			}
		}
	}()
	return tracker, nil
}
