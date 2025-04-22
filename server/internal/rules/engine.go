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
	"time"

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
		return
	}

	for _, rule := range activeRules {
		for _, m := range metrics {
			if !ruleMatches(rule, m, meta) {
				continue
			}

			// Track time window of matching metrics
			key := rule.ID + "|" + meta.EndpointID
			e.history[key] = append(e.history[key], m)

			// Trim old metrics based on rule duration
			dur, err := time.ParseDuration(rule.Trigger.Duration)
			if err != nil {
				continue
			}
			e.history[key] = trimOldMetrics(e.history[key], dur)

			// Evaluate condition (e.g., > threshold for N seconds)
			if len(e.history[key]) == 0 {
				continue
			}

			triggered := true
			for _, sample := range e.history[key] {
				switch rule.Trigger.Operator {
				case "gt":
					if sample.Value <= rule.Trigger.Threshold {
						triggered = false
					}
				case "lt":
					if sample.Value >= rule.Trigger.Threshold {
						triggered = false
					}
				case "eq":
					if sample.Value != rule.Trigger.Threshold {
						triggered = false
					}
				}
				if !triggered {
					break
				}
			}

			if triggered {
				if !e.firing[key] {
					e.alertMgr.HandleState(ctx, rule, meta, m.Value, triggered)
				}

			} else {
				delete(e.firing, key)

			}
		}
	}
}

func trimOldMetrics(metrics []model.Metric, window time.Duration) []model.Metric {
	cutoff := time.Now().Add(-window)
	var trimmed []model.Metric
	for _, m := range metrics {
		if m.Timestamp.After(cutoff) {
			trimmed = append(trimmed, m)
		}
	}
	return trimmed
}
