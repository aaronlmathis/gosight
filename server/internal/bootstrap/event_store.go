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

// gosight/agent/internal/bootstrap/user_store.go
// // Package bootstrap initializes the user store
// Package store provides an interface for storing and retrieving user / permission data

package bootstrap

import (
	"context"
	"errors"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store/eventstore"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// In this file, we initialize the event store for the GoSight server.
// The event store is responsible for storing and retrieving events.
// The implementation of this interface should handle the actual storage
// and retrieval of events, whether it's in memory, a database, or any other
// storage mechanism.

func InitEventStore(ctx context.Context, cfg *config.Config) (eventstore.EventStore, error) {
	utils.Info("Initializing user store type: %s", cfg.EventStore.Engine)
	switch cfg.EventStore.Engine {
	case "memory":
		return eventstore.NewMemoryStore(500)
	case "json":
		return eventstore.NewJSONEventStore(cfg.EventStore.Path, 500)

	default:
		return nil, errors.New("invalid event backend")
	}
}
