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
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/aaronlmathis/gosight/server/internal/alerts"
	"github.com/aaronlmathis/gosight/server/internal/store/rulestore"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type Evaluator struct {
	store    rulestore.RuleStore
	AlertMgr *alerts.Manager
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
		AlertMgr: alertMgr,
		history:  make(map[string][]model.Metric),
		firing:   make(map[string]bool), // ruleID + endpointID
	}
}

// EvaluateMetric processes the given metrics and metadata,
// checking them against active rules in the store.
// It emits events when rules are triggered based on the metrics.
// The evaluation is done in the context of the provided context.Context.
// The metrics are expected to be in the format of model.Metric,
// and the metadata is expected to be in the format of model.Meta.
func (e *Evaluator) EvaluateMetric(ctx context.Context, metrics []model.Metric, meta *model.Meta) {
	//utils.Debug("Evaluating metrics for rules...")

	activeRules, err := e.store.GetActiveRules(ctx)
	if err != nil {
		utils.Error("Failed to fetch active rules: %v", err)
		return
	}

	for _, rule := range activeRules {
		if rule.Type != "metric" { // Only apply metric-type rules
			continue
		}
		//utils.Debug("Checking rule %s with expression %s", rule.ID, rule.Expression)
		if !ruleMatchLabels(rule.Match, meta) {
			//utils.Debug("Rule %s did not match endpoint %s or tags", rule.ID, meta.EndpointID)
			continue
		}

		// Build map[string]interface{} for expression
		values := make(map[string]interface{})
		for _, m := range metrics {
			// Example: "mem_used_percent" = 72.5
			key := strings.ToLower(m.SubNamespace + "_" + m.Name)
			values[key] = m.Value
			// Dimensions
			for k, v := range m.Dimensions {
				values[k] = v
			}
		}

		// Meta tags
		for k, v := range meta.Tags {
			values[k] = v
		}

		//fmt.Printf("Expression to parse: '%s'\n", rule.Expression)
		expression, err := govaluate.NewEvaluableExpression(rule.Expression)

		if err != nil {
			//fmt.Printf("Expression parse failed for rule [%s]: %v", rule.ID, err)
			continue
		}

		result, err := expression.Evaluate(values)
		if err != nil {
			//fmt.Printf("Expression evaluate failed for rule %s: %v", rule.ID, err)
			continue
		}

		//utils.Debug("→ Expression result for rule %s: %v", rule.ID, result)

		isTriggered, _ := result.(bool)
		key := rule.ID + "|" + meta.EndpointID

		if isTriggered {
			if !e.firing[key] {
				e.firing[key] = true
				e.AlertMgr.HandleState(ctx, rule, meta, extractFirstValue(values), true)
			}
		} else {
			if e.firing[key] {
				delete(e.firing, key)
				e.AlertMgr.HandleState(ctx, rule, meta, extractFirstValue(values), false)
			}
		}
	}
}

// EvaluateLogs processes the given logs and metadata,
// checking them against active rules in the store.
// It emits events when rules are triggered based on the logs.
// The evaluation is done in the context of the provided context.Context.
// The logs are expected to be in the format of model.LogEntry,
// and the metadata is expected to be in the format of model.Meta.
// Logs are point-in-time events, so they are always evaluated immediately.
func (e *Evaluator) EvaluateLogs(ctx context.Context, logs []model.LogEntry, meta *model.Meta) {
	activeRules, err := e.store.GetActiveRules(ctx)
	if err != nil {
		utils.Error("Failed to fetch active rules: %v", err)
		return
	}

	for _, log := range logs {
		for _, rule := range activeRules {
			if rule.Type != "log" { // Only apply log-type rules
				continue
			}

			if !ruleMatchLabels(rule.Match, meta) {
				continue
			}

			values := make(map[string]interface{})

			// Map log fields
			values["level"] = log.Level
			values["message"] = log.Message
			values["source"] = log.Source
			values["category"] = log.Category

			// Map log tags
			for k, v := range log.Tags {
				values[k] = v
			}

			// Map log Fields
			for k, v := range log.Fields {
				values[k] = v
			}

			// Map endpoint meta tags
			for k, v := range meta.Tags {
				values[k] = v
			}

			expression, err := govaluate.NewEvaluableExpressionWithFunctions(rule.Expression, map[string]govaluate.ExpressionFunction{
				"contains": func(args ...interface{}) (interface{}, error) {
					if len(args) != 2 {
						return false, nil
					}
					str, ok1 := args[0].(string)
					substr, ok2 := args[1].(string)
					if !ok1 || !ok2 {
						return false, nil
					}
					return strings.Contains(str, substr), nil
				},
			})
			if err != nil {
				utils.Warn("Invalid log rule expression %s: %v", rule.ID, err)
				continue
			}

			result, err := expression.Evaluate(values)
			if err != nil {
				utils.Warn("Failed to evaluate log rule %s: %v", rule.ID, err)
				continue
			}

			isTriggered, _ := result.(bool)

			if isTriggered {
				// Always fire immediately for logs (logs are point-in-time events)
				e.AlertMgr.HandleLogState(ctx, rule, meta, log, true)
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
