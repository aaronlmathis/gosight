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

// gosight/agent/internal/store/eventstore/memory.go

// Package eventstore provides an interface for storing and retrieving events.
package eventstore

import (
	"sync"

	"github.com/aaronlmathis/gosight/shared/model"
)

// MemoryEventStore is an in-memory implementation of the EventStore interface.
type MemoryEventStore struct {
	lock   sync.RWMutex
	events []model.EventEntry
	max    int
}

// NewMemoryStore creates a new MemoryEventStore with a specified maximum number of entries.
func NewMemoryStore(maxEntries int) (*MemoryEventStore, error) {
	return &MemoryEventStore{
		events: make([]model.EventEntry, 0, maxEntries),
		max:    maxEntries,
	}, nil
}

// AddEvent adds a new event entry to the store.
// If the store is full, it drops the oldest event to make room for the new one.
// This method is thread-safe and uses a write lock to ensure that only one
// goroutine can modify the store at a time.

func (s *MemoryEventStore) AddEvent(e model.EventEntry) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.events) >= s.max {
		s.events = s.events[1:] // drop oldest
	}
	s.events = append(s.events, e)
}

// GetRecent retrieves the most recent event entries from the store.
// It returns a slice of EventEntry objects, limited to the specified number.
// If the limit exceeds the number of stored events, it returns all available events.
// This method is thread-safe and uses a read lock to allow multiple goroutines
// to read from the store concurrently.

func (s *MemoryEventStore) GetRecent(limit int) []model.EventEntry {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if limit > len(s.events) {
		limit = len(s.events)
	}
	return append([]model.EventEntry(nil), s.events[len(s.events)-limit:]...)
}
