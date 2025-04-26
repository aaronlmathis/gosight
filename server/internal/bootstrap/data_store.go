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

// gosight/agent/internal/bootstrap/data_store.go
// // Package bootstrap initializes the user store
// Package store provides an interface for storing and retrieving user / permission data

package bootstrap

import (
	"database/sql"
	"fmt"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store/datastore"
	pgstore "github.com/aaronlmathis/gosight/server/internal/store/datastore/pgsql"

	"github.com/aaronlmathis/gosight/shared/utils"
	_ "github.com/lib/pq"
)

func InitDataStore(cfg *config.Config) (datastore.DataStore, error) {
	dataStoreType := cfg.DataStore.Engine

	utils.Info("Initializing user store type: %s", dataStoreType)
	var dataStore datastore.DataStore
	switch cfg.DataStore.Engine {
	case "postgres":
		db, err := sql.Open("postgres", cfg.DataStore.DSN) // TODO 	Fix this to be more generic
		if err != nil {
			return nil, err
		}

		// Optionally test the connection
		if err := db.Ping(); err != nil {
			return nil, err
		}
		dataStore = pgstore.NewPGDataStore(db)

	default:
		return nil, fmt.Errorf("unsupported datastore type: %s", dataStoreType)
	}

	utils.Info("Data store [%s] initialized successfully", dataStoreType)
	return dataStore, nil
}
