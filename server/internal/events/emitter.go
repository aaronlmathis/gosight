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

// gosight/agent/shared/model/event.go
// Package events provides an event emitter for the GoSight application.
package events

import (
	"context"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/store/eventstore"
	"github.com/aaronlmathis/gosight/server/internal/websocket"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// Emitter is an event emitter that stores events in an event store.
// It provides a method to emit events with various attributes such as level, category, message, source, and metadata.
type Emitter struct {
	Store eventstore.EventStore
	hub   *websocket.Hub
}

// NewEmitter creates a new Emitter instance with the provided event store.
func NewEmitter(store eventstore.EventStore, hub *websocket.Hub) *Emitter {
	return &Emitter{
		Store: store,
		hub:   hub,
	}
}

// Emit emits an event with the specified attributes.
func (e *Emitter) Emit(ctx context.Context, event model.EventEntry) {
	if event.ID == "" {
		event.ID = model.GenerateID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	if e.hub != nil {
		utils.Debug("Emmitter broadcasting event: %s", event.ID)
		e.hub.BroadcastEvent(event)
	}
	e.Store.AddEvent(ctx, event)
}
