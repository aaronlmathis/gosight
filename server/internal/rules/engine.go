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
	"fmt"
	"regexp"
	"strconv"
	"strings"

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
		firing:   make(map[string]bool),
	}
}

// EvaluateMetric processes the given metrics and metadata,
// checking them against active rules in the store.
// It emits events when rules are triggered based on the metrics.
// The evaluation is done in the context of the provided context.Context.
// The metrics are expected to be in the format of model.Metric,
// and the metadata is expected to be in the format of model.Meta.
func (e *Evaluator) EvaluateMetric(ctx context.Context, metrics []model.Metric, meta *model.Meta) {

	activeRules, err := e.store.GetActiveRules(ctx)
	if err != nil {
		utils.Error("Failed to fetch active rules: %v", err)
		return
	}

	for _, rule := range activeRules {
		utils.Debug("Evaluating rule: %s", rule.ID)
		if !rule.Enabled {
			continue
		}

		if rule.Type != "metric" {
			continue
		}
		utils.Debug("Before label check")
		if !ruleMatchLabels(rule.Match, meta) {
			continue
		}
		utils.Debug("After label check")
		metricName := fmt.Sprintf("%s.%s.%s", rule.Scope.Namespace, rule.Scope.SubNamespace, rule.Scope.Metric)

		utils.Debug("Metric name to match: %s", metricName)
		var matched *model.Metric
		for _, m := range metrics {
			full := strings.ToLower(fmt.Sprintf("%s.%s.%s", m.Namespace, m.SubNamespace, m.Name))
			utils.Debug("Checking metric: %s against rule metric: %s", full, metricName)
			if full == metricName {
				matched = &m
				break
			}
		}
		utils.Debug("After ranging metric names...")
		if matched == nil {
			continue
		}

		utils.Debug("About to fire rule: %s", rule.ID)

		firing := evaluateExpression(rule.Expression, matched)
		key := rule.ID + "|" + meta.EndpointID

		if firing {
			if !e.firing[key] {
				e.firing[key] = true

				e.AlertMgr.HandleState(ctx, rule, meta, matched.Value, true)
			}
		} else {
			if e.firing[key] {
				delete(e.firing, key)
				e.AlertMgr.HandleState(ctx, rule, meta, matched.Value, false)
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
	utils.Debug("✅  Evaluate Metric called.")
	activeRules, err := e.store.GetActiveRules(ctx)
	if err != nil {
		utils.Error("Failed to fetch active rules: %v", err)
		return
	}

	for _, log := range logs {
		for _, rule := range activeRules {
			utils.Debug("Evaluating log rule: %s", rule.ID)
			if !rule.Enabled {
				continue
			}
			if rule.Type != "log" {
				continue
			}
			if !ruleMatchLabels(rule.Match, meta) {
				continue
			}
			if rule.Match.Category != "" && rule.Match.Category != log.Category {
				continue
			}
			if rule.Match.Source != "" && rule.Match.Source != log.Source {
				continue
			}
			firing := evaluateLogExpression(rule.Expression, log)
			if firing {
				utils.Debug("✅  Firing alert for rule: %s, endpoint: %s", rule.ID, meta.EndpointID)
				e.AlertMgr.HandleLogState(ctx, rule, meta, log, true)
			}
		}
	}
}

// evaluateLogExpression evaluates the log entry against the rule's expression.
func evaluateLogExpression(expr model.Expression, log model.LogEntry) bool {
	val := ""
	if expr.Datatype == "level" {
		val = log.Level
	} else if expr.Datatype == "message" {
		val = log.Message
	} else {
		val = log.Source
	}

	switch expr.Operator {
	case "contains":
		return strings.Contains(val, toString(expr.Value))
	case "=", "==":
		return val == toString(expr.Value)
	case "!=":
		return val != toString(expr.Value)
	case "regex":
		re, err := regexp.Compile(toString(expr.Value))
		if err != nil {
			return false
		}
		return re.MatchString(val)
	default:
		return false
	}
}

// evaluateLogExpression evaluates the log entry against the rule's expression.
func evaluateExpression(expr model.Expression, m *model.Metric) bool {
	switch expr.Operator {
	case ">":
		return m.Value > toFloat(expr.Value)
	case "<":
		return m.Value < toFloat(expr.Value)
	case ">=":
		return m.Value >= toFloat(expr.Value)
	case "<=":
		return m.Value <= toFloat(expr.Value)
	case "=", "==":
		return m.Value == toFloat(expr.Value)
	case "!=":
		return m.Value != toFloat(expr.Value)
	case "contains":
		return strings.Contains(toString(m.Value), toString(expr.Value))
	case "regex":
		re, err := regexp.Compile(toString(expr.Value))
		if err != nil {
			return false
		}
		return re.MatchString(toString(m.Value))
	default:
		return false
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

	for k, v := range match.Labels {
		if meta.Tags[k] != v {
			return false
		}
	}
	return true
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0.0
	}
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.2f", val)
	case int:
		return fmt.Sprintf("%d", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
