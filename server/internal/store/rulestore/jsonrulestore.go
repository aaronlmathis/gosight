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
	"encoding/json"
	"os"
	"sync"

	"github.com/aaronlmathis/gosight/shared/model"
)

// JSONRuleStore is a rule store that uses a JSON file for persistence.
// It implements the RuleStore interface and provides methods for adding,
type JSONRuleStore struct {
	path  string
	rules map[string]model.AlertRule
	lock  sync.RWMutex
}

// NewJSONStore creates a new JSONRuleStore with the specified file path.
// It loads existing rules from the file if it exists.
func NewJSONStore(path string) (*JSONRuleStore, error) {
	j := &JSONRuleStore{
		path:  path,
		rules: make(map[string]model.AlertRule),
	}
	_ = j.load()
	return j, nil
}

// load reads the rules from the JSON file and populates the rules map.
// If the file does not exist, it initializes an empty rules map.

func (j *JSONRuleStore) load() error {
	data, err := os.ReadFile(j.path)
	if err != nil {
		return err
	}
	var list []model.AlertRule
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	for _, r := range list {
		j.rules[r.ID] = r
	}
	return nil
}

// save writes the current rules to the JSON file.
// It is called after any modification to the rules map.

func (j *JSONRuleStore) save() {
	data, _ := json.MarshalIndent(j.rules, "", "  ")
	_ = os.WriteFile(j.path, data, 0644)
}

// AddRule adds a new rule to the store.
// It locks the store for writing, adds the rule, and then saves the rules to the file.
func (j *JSONRuleStore) AddRule(ctx context.Context, r model.AlertRule) error {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.rules[r.ID] = r
	j.save()
	return nil
}

// UpdateRule updates an existing rule in the store.
// It locks the store for writing, updates the rule, and then saves the rules to the file.
func (j *JSONRuleStore) UpdateRule(ctx context.Context, r model.AlertRule) error {
	return j.AddRule(ctx, r)
}

// DeleteRule removes a rule from the store.
// It locks the store for writing, deletes the rule, and then saves the rules to the file.
func (j *JSONRuleStore) DeleteRule(ctx context.Context, id string) error {
	j.lock.Lock()
	defer j.lock.Unlock()
	delete(j.rules, id)
	j.save()
	return nil
}

// ListRules returns a list of all rules in the store.
// It locks the store for reading and returns a slice of AlertRule.

func (j *JSONRuleStore) ListRules(ctx context.Context) ([]model.AlertRule, error) {
	j.lock.RLock()
	defer j.lock.RUnlock()

	var list []model.AlertRule
	for _, r := range j.rules {
		list = append(list, r)
	}
	return list, nil
}

// GetActiveRules returns a list of all active rules in the store.
// It calls ListRules to get all rules and filters them based on the Enabled field.
func (j *JSONRuleStore) GetActiveRules(ctx context.Context) ([]model.AlertRule, error) {
	all, _ := j.ListRules(ctx)
	var active []model.AlertRule
	for _, r := range all {
		if r.Enabled {
			active = append(active, r)
		}
	}
	return active, nil
}

// GetRuleByID retrieves a rule by its ID from the JSON rule store.
func (s *JSONRuleStore) GetRuleByID(ctx context.Context, id string) (model.AlertRule, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if rule, ok := s.rules[id]; ok {
		return rule, nil
	}
	return model.AlertRule{}, os.ErrNotExist
}

// GetRuleByName retrieves a rule by its Name from the JSON rule store.
func (s *JSONRuleStore) GetRuleByName(ctx context.Context, name string) (model.AlertRule, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, rule := range s.rules {
		if rule.Name == name {
			return rule, nil
		}
	}
	return model.AlertRule{}, os.ErrNotExist
}
