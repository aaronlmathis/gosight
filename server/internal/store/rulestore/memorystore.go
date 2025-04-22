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

package rulestore

import (
	"context"
	"sync"

	"github.com/aaronlmathis/gosight/shared/model"
)

// internal/store/rulestore/memorystore.go

// Package memorystore provides an in-memory implementation of the RuleStore interface.
// It is primarily used for testing and development purposes.

type MemoryRuleStore struct {
	rules map[string]model.AlertRule
	lock  sync.RWMutex
}

func NewMemoryStore() *MemoryRuleStore {
	return &MemoryRuleStore{rules: make(map[string]model.AlertRule)}
}

func (s *MemoryRuleStore) AddRule(ctx context.Context, r model.AlertRule) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.rules[r.ID] = r
	return nil
}

func (s *MemoryRuleStore) UpdateRule(ctx context.Context, r model.AlertRule) error {
	return s.AddRule(ctx, r)
}

func (s *MemoryRuleStore) DeleteRule(ctx context.Context, id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.rules, id)
	return nil
}

func (s *MemoryRuleStore) ListRules(ctx context.Context) ([]model.AlertRule, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var out []model.AlertRule
	for _, rule := range s.rules {
		out = append(out, rule)
	}
	return out, nil
}

func (s *MemoryRuleStore) GetActiveRules(ctx context.Context) ([]model.AlertRule, error) {
	all, _ := s.ListRules(ctx)
	var filtered []model.AlertRule
	for _, r := range all {
		if r.Enabled {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}
