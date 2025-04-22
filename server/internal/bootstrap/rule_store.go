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

// gosight/agent/internal/bootstrap/rule_store.go
// // Package bootstrap initializes the rule store

package bootstrap

import (
	"errors"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store/rulestore"
)

// InitRuleStore initializes the rule store based on the configuration.
func InitRuleStore(cfg *config.Config) (rulestore.RuleStore, error) {
	switch cfg.RuleStore.Engine {
	case "yaml":
		return rulestore.NewYAMLStore(cfg.RuleStore.Path)
	//case "memory":
	//return rulestore.NewMemoryStore(), nil
	default:
		return nil, errors.New("unsupported rule store engine: " + cfg.RuleStore.Engine)
	}
}
