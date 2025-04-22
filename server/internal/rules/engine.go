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

// Package rules provides the core logic for evaluating alert rules
// and emitting events based on metric data.
package rules

import (
	"context"

	"github.com/Knetic/govaluate"
	"github.com/aaronlmathis/gosight/server/internal/alerts"
	"github.com/aaronlmathis/gosight/server/internal/store/rulestore"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type Evaluator struct {
	store    rulestore.RuleStore
	alertMgr *alerts.Manager
	history  map[string][]model.Metric
	firing   map[string]bool // ruleID + endpointID
}

// NewEvaluator creates a new Evaluator instance.
// It takes a RuleStore for rule management and an *events.emitter
// for emitting events. The history map is used to track metrics
// for each rule and endpoint combination.
// The history is keyed by rule ID and endpoint ID, allowing for
// efficient tracking of metrics over time.
func NewEvaluator(store rulestore.RuleStore, alertMgr *alerts.Manager) *Evaluator {
	return &Evaluator{
		store:    store,
		alertMgr: alertMgr,
		history:  make(map[string][]model.Metric),
		firing:   make(map[string]bool), // ruleID + endpointID
	}
}

// Evaluate processes the given metrics and metadata,
// checking them against active rules in the store.
// It emits events when rules are triggered based on the metrics.
// The evaluation is done in the context of the provided context.Context.
// The metrics are expected to be in the format of model.Metric,
// and the metadata is expected to be in the format of model.Meta.
func (e *Evaluator) Evaluate(ctx context.Context, metrics []model.Metric, meta *model.Meta) {
	utils.Debug("Evaluating metrics for rules...")

	activeRules, err := e.store.GetActiveRules(ctx)
	if err != nil {
		utils.Error("Failed to fetch active rules: %v", err)
		return
	}

	for _, rule := range activeRules {
		if !ruleMatchLabels(rule.Match, meta) {
			continue
		}

		// Build map[string]interface{} for expression
		values := make(map[string]interface{})
		for _, m := range metrics {
			// Example: "mem.used_percent" = 72.5
			key := m.SubNamespace + "." + m.Name
			values[key] = m.Value
		}

		// Evaluate expression
		result, err := govaluate.NewEvaluableExpression(rule.Expression)
		if err != nil {
			utils.Error("Invalid rule expression (%s): %v", rule.ID, err)
			continue
		}

		ok, err := result.Evaluate(values)
		if err != nil {
			utils.Error("Evaluation failed for rule (%s): %v", rule.ID, err)
			continue
		}

		isTriggered, _ := ok.(bool)
		key := rule.ID + "|" + meta.EndpointID

		if isTriggered {
			if !e.firing[key] {
				e.firing[key] = true
				e.alertMgr.HandleState(ctx, rule, meta, extractFirstValue(values), true)
			}
		} else {
			if e.firing[key] {
				delete(e.firing, key)
				e.alertMgr.HandleState(ctx, rule, meta, extractFirstValue(values), false)
			}
		}
	}
}

// ruleMatchLabels checks if the rule's match labels match the given metadata labels.
func ruleMatchLabels(match model.MatchCriteria, meta *model.Meta) bool {
	if len(match.EndpointIDs) > 0 {
		found := false
		for _, id := range match.EndpointIDs {
			if id == meta.EndpointID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for k, v := range match.TagSelectors {
		if meta.Tags[k] != v {
			return false
		}
	}

	return true
}

// extractFirstValue extracts the first float64 value from the map.
// It returns 0.0 if no float64 value is found.
func extractFirstValue(values map[string]interface{}) float64 {
	for _, val := range values {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}
