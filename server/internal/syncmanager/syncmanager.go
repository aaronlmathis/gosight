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

// gosight/internal/syncmanager/syncmanager.go

// SyncManager is responsible for managing synchronization between the cache and the datastore.

package syncmanager

import (
	"context"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/cache"
	"github.com/aaronlmathis/gosight/server/internal/store/datastore"
	"github.com/aaronlmathis/gosight/server/internal/tracker"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// SyncManager handles periodic persistence of in-memory caches to the database.
type SyncManager struct {
	ctx       context.Context
	Cache     *cache.Cache
	DataStore datastore.DataStore
	Tracker   *tracker.EndpointTracker
	Interval  time.Duration
	wg        sync.WaitGroup
}

// New creates a new SyncManager.
func NewSyncManager(ctx context.Context, c *cache.Cache, ds datastore.DataStore, tracker *tracker.EndpointTracker, interval time.Duration) *SyncManager {
	return &SyncManager{
		ctx:       ctx,
		Cache:     c,
		DataStore: ds,
		Tracker:   tracker,
		Interval:  interval,
	}
}

// Run starts the sync manager and blocks until ctx is cancelled. On shutdown, it flushes all caches.
func (s *SyncManager) Run() {
	s.wg.Add(3)
	go s.runTagSync()
	go s.runEndpointTrackerSync()
	go s.runLifecycleEmitter()

	<-s.ctx.Done()
	utils.Info("[syncer] context canceled, flushing all caches once")
	s.SyncOnce()

	s.wg.Wait()
	utils.Info("[syncer] all syncers shut down cleanly")
}

func (s *SyncManager) runTagSync() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()

	utils.Info("[syncer] tag flush started")
	for {
		select {
		case <-ticker.C:
			utils.Info("[syncer] flushing tag cache")
			s.Cache.Tags.Flush(s.DataStore)
		case <-s.ctx.Done():
			utils.Info("[syncer] tag sync stopped")
			return
		}
	}
}

func (s *SyncManager) runEndpointTrackerSync() {

	defer s.wg.Done()
	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()

	utils.Info("[syncer] endpoint flush started")
	for {
		select {
		case <-ticker.C:
			s.Tracker.SyncToStore(s.ctx, s.DataStore)
		case <-s.ctx.Done():
			utils.Info("[syncer] endpoint tracker sync stopped")
			return
		}
	}

}

func (s *SyncManager) runLifecycleEmitter() {

	defer s.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	utils.Info("[syncer] lifecycle emitter started")
	for {
		select {
		case <-ticker.C:
			s.Tracker.CheckAgentStatusesAndEmitEvents()
			s.Tracker.CheckContainerStatusesAndEmit()
		case <-s.ctx.Done():
			utils.Info("[syncer] endpoint tracker lifecycle loop stopped")
			return
		}
	}

}

// SyncOnce triggers a one-time flush for all caches. Useful for shutdown or CLI tools.
func (s *SyncManager) SyncOnce() {
	utils.Info("[syncer] flushing all caches ONCE")
	s.Cache.Tags.Flush(s.DataStore)
	s.Tracker.SyncToStore(context.Background(), s.DataStore) // use ctx background because by the time this is called, the main context may be canceled
	// Add additional flush calls here (e.g. s.Cache.Processes.Flush(...))
}
