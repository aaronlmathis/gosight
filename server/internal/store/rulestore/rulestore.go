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

// internal/store/rulestore/rulestore.go
// Package rulestore provides an interface for managing alert rules in the system.
// It includes methods for adding, updating, deleting, and retrieving rules.

package rulestore

import (
	"context"

	"github.com/aaronlmathis/gosight/shared/model"
)

// RuleStore defines the interface for managing alert rules.
type RuleStore interface {
	AddRule(ctx context.Context, rule model.AlertRule) error
	UpdateRule(ctx context.Context, rule model.AlertRule) error
	DeleteRule(ctx context.Context, id string) error
	ListRules(ctx context.Context) ([]model.AlertRule, error)
	GetActiveRules(ctx context.Context) ([]model.AlertRule, error)
}
