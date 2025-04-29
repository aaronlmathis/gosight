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

// gosight/agent/internal/bootstrap/alert_store.go

package bootstrap

import (
	"database/sql"
	"fmt"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store/alertstore"
	"github.com/aaronlmathis/gosight/server/internal/store/alertstore/pgalertstore"

	"github.com/aaronlmathis/gosight/shared/utils"
	_ "github.com/lib/pq"
)

// InitAlertStore initializes the alert store for the GoSight server.
// The alert store is responsible for storing and retrieving alert instances.

func InitAlertStore(cfg *config.Config) (alertstore.AlertStore, error) {
	alertStoreType := cfg.AlertStore.Engine

	utils.Info("Initializing alert store type: %s", alertStoreType)
	var alertStore alertstore.AlertStore
	switch cfg.AlertStore.Engine {
	case "postgres":
		db, err := sql.Open("postgres", cfg.AlertStore.DSN)
		if err != nil {
			return nil, err
		}

		// Optionally test the connection
		if err := db.Ping(); err != nil {
			return nil, err
		}
		alertStore = pgalertstore.NewPGAlertStore(db)

	default:
		return nil, fmt.Errorf("unsupported alert store type: %s", alertStoreType)
	}

	utils.Info("Alert store [%s] initialized successfully", alertStoreType)
	return alertStore, nil
}
