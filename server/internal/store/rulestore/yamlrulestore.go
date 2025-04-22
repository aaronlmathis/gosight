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
	"os"
	"sync"

	"github.com/aaronlmathis/gosight/shared/model"
	"gopkg.in/yaml.v3"
)

// YAMLRuleStore is a rule store that uses a YAML file for persistence.
// It implements the RuleStore interface and provides methods for adding,
// updating, deleting, and retrieving rules.
type YAMLRuleStore struct {
	path  string
	lock  sync.RWMutex
	rules map[string]model.AlertRule
}

// NewYAMLStore creates a new YAMLRuleStore with the specified file path.
func NewYAMLStore(path string) (*YAMLRuleStore, error) {
	s := &YAMLRuleStore{
		path:  path,
		rules: make(map[string]model.AlertRule),
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// load reads the rules from the YAML file and populates the rules map.
// If the file does not exist, it initializes an empty rules map.
func (s *YAMLRuleStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	var list []model.AlertRule
	if err := yaml.Unmarshal(data, &list); err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	for _, r := range list {
		s.rules[r.ID] = r
	}
	return nil
}

// save writes the current rules to the YAML file.
func (s *YAMLRuleStore) save() error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var list []model.AlertRule
	for _, r := range s.rules {
		list = append(list, r)
	}

	data, err := yaml.Marshal(list)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

// AddRule adds a new rule to the store.
func (s *YAMLRuleStore) AddRule(ctx context.Context, r model.AlertRule) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.rules[r.ID] = r
	return s.save()
}

// UpdateRule updates an existing rule in the store.
func (s *YAMLRuleStore) UpdateRule(ctx context.Context, r model.AlertRule) error {
	return s.AddRule(ctx, r)
}

// DeleteRule removes a rule from the store.
func (s *YAMLRuleStore) DeleteRule(ctx context.Context, id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.rules, id)
	return s.save()
}

// ListRules returns a list of all rules in the store.
func (s *YAMLRuleStore) ListRules(ctx context.Context) ([]model.AlertRule, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var out []model.AlertRule
	for _, r := range s.rules {
		out = append(out, r)
	}
	return out, nil
}

// GetActiveRules returns a list of active rules in the store.
// It filters the rules based on their Enabled status.

func (s *YAMLRuleStore) GetActiveRules(ctx context.Context) ([]model.AlertRule, error) {
	all, _ := s.ListRules(ctx)
	var filtered []model.AlertRule
	for _, r := range all {
		if r.Enabled {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

// GetRuleByID retrieves a rule by its ID.
func (s *YAMLRuleStore) GetRuleByID(ctx context.Context, id string) (model.AlertRule, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if rule, ok := s.rules[id]; ok {
		return rule, nil
	}
	return model.AlertRule{}, os.ErrNotExist
}

// GetRuleByName retrieves a rule by its Name (case-sensitive).
func (s *YAMLRuleStore) GetRuleByName(ctx context.Context, name string) (model.AlertRule, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, rule := range s.rules {
		if rule.Name == name {
			return rule, nil
		}
	}
	return model.AlertRule{}, os.ErrNotExist
}
