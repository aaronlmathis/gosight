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
// server/internal/store/alertstore/alertstore.go

package alertstore

import (
	"context"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
)

// AlertStore is an interface for managing alert instances in the database.
// It provides methods to upsert, resolve, and list active and historical alerts.
// The interface is designed to be implemented by different database backends,
// allowing for flexibility in the storage solution used by the application.

type AlertStore interface {
	UpsertAlert(ctx context.Context, a *model.AlertInstance) error
	ResolveAlert(ctx context.Context, ruleID, target string, resolvedAt time.Time) error
	ListActiveAlerts(ctx context.Context) ([]model.AlertInstance, error)
	ListAlertHistory(ctx context.Context, since time.Time) ([]model.AlertInstance, error)
}
