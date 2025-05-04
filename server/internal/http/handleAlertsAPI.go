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

// server/internal/http/handleAlertsAPI.go

package httpserver

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/google/uuid"
)

// HandleAlertsAPI handles requests to the /api/alerts endpoint
// It retrieves all alerts from the database and returns them as JSON.

func (s *HttpServer) HandleAlertsAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	limit := utils.ParseIntOrDefault(query.Get("limit"), 100)
	page := utils.ParseIntOrDefault(query.Get("page"), 1)

	alertQuery := model.AlertQuery{
		RuleID: strings.TrimSpace(query.Get("rule_id")),
		State:  strings.TrimSpace(query.Get("state")),
		Level:  strings.TrimSpace(query.Get("level")),
		Target: strings.TrimSpace(query.Get("target")),
		Scope:  strings.TrimSpace(query.Get("scope")),
		Sort:   strings.TrimSpace(query.Get("sort")),
		Order:  strings.TrimSpace(query.Get("order")),
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	// Total count
	total, err := s.Sys.Stores.Alerts.CountAlertsFiltered(ctx, alertQuery)
	if err != nil {
		http.Error(w, "failed to count alerts", http.StatusInternalServerError)
		return
	}

	// Filtered + paged alerts
	alerts, err := s.Sys.Stores.Alerts.ListAlertsFiltered(ctx, alertQuery)
	if err != nil {
		http.Error(w, "failed to query alerts", http.StatusInternalServerError)
		return
	}

	tags := make(map[string]string)
	for _, raw := range query["tag"] {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) == 2 {
			tags[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	alertQuery.Tags = tags

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleActiveAlertsAPI handles requests to the /api/alerts/active endpoint

func (s *HttpServer) HandleActiveAlertsAPI(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	alerts := s.Sys.Tele.Alerts.ListActive()
	if alerts == nil {
		alerts = []model.AlertInstance{}
	}

	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		http.Error(w, "failed to encode alerts", http.StatusInternalServerError)
		return
	}
}

// HandleAlertRulesAPI handles requests to the /api/alerts/rules endpoint

func (s *HttpServer) HandleAlertRulesAPI(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	alertRules, err := s.Sys.Stores.Rules.ListRules(r.Context())
	if err != nil {
		http.Error(w, "failed to load alert rules", http.StatusInternalServerError)
		return
	}

	if alertRules == nil {
		alertRules = []model.AlertRule{}
	}

	if err := json.NewEncoder(w).Encode(alertRules); err != nil {
		http.Error(w, "failed to encode alert rules", http.StatusInternalServerError)
		return
	}
}

// HandleAlertsSummaryAPI handles requests to the /api/alerts/summary endpoint
// It returns a summary of alerts grouped by rule_id, showing the latest state and last change time.
func (s *HttpServer) HandleAlertsSummaryAPI(w http.ResponseWriter, r *http.Request) {
	alerts, err := s.Sys.Stores.Alerts.ListAlerts(r.Context())
	if err != nil {
		http.Error(w, "failed to fetch alerts", http.StatusInternalServerError)
		return
	}

	// Build summary: latest state per rule_id
	summaryMap := make(map[string]model.AlertSummary)

	for _, a := range alerts {
		existing, exists := summaryMap[a.RuleID]
		if !exists || a.LastFired.After(existing.LastChange) {
			summaryMap[a.RuleID] = model.AlertSummary{
				RuleID:     a.RuleID,
				State:      a.State,
				LastChange: a.LastFired,
			}
		}
	}

	var summaries []model.AlertSummary
	for _, v := range summaryMap {
		summaries = append(summaries, v)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summaries); err != nil {
		http.Error(w, "failed to encode summary", http.StatusInternalServerError)
		return
	}
}

// HandleCreateAlert handles POST /api/v1/alerts
func (s *HttpServer) HandleCreateAlertRuleAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var rule model.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		utils.Error("Failed to decode alert rule: %v", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Generate ID if missing
	if strings.TrimSpace(rule.ID) == "" {
		rule.ID = uuid.NewString()
	}

	// Name must be unique
	existing, err := s.Sys.Stores.Rules.GetRuleByName(ctx, rule.Name)
	if err == nil && existing.ID != "" {
		utils.Warn("Duplicate rule name: %s", rule.Name)
		http.Error(w, "rule name already exists", http.StatusConflict)
		return
	}

	if err := s.Sys.Stores.Rules.AddRule(ctx, rule); err != nil {
		utils.Error("Failed to save rule: %v", err)
		http.Error(w, "failed to save rule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"id":     rule.ID,
	})
}

func SortBy(alerts *[]model.AlertInstance, field, order string) {
	slice := *alerts
	field = strings.ToLower(strings.TrimSpace(field))
	order = strings.ToLower(strings.TrimSpace(order))

	less := func(i, j int) bool { return true }

	switch field {
	case "first_fired":
		less = func(i, j int) bool { return slice[i].FirstFired.Before(slice[j].FirstFired) }
	case "last_ok":
		less = func(i, j int) bool { return slice[i].LastOK.Before(slice[j].LastOK) }
	case "rule_id":
		less = func(i, j int) bool { return slice[i].RuleID < slice[j].RuleID }
	case "target":
		less = func(i, j int) bool { return slice[i].Target < slice[j].Target }
	case "scope":
		less = func(i, j int) bool { return slice[i].Scope < slice[j].Scope }
	case "state":
		less = func(i, j int) bool { return slice[i].State < slice[j].State }
	case "level":
		less = func(i, j int) bool { return slice[i].Level < slice[j].Level }
	case "last_fired":
		less = func(i, j int) bool {
			var t1, t2 time.Time
			if !slice[i].LastFired.IsZero() {
				t1 = slice[i].LastFired
			} else {
				t1 = slice[i].FirstFired
			}
			if !slice[j].LastFired.IsZero() {
				t2 = slice[j].LastFired
			} else {
				t2 = slice[j].FirstFired
			}
			return t1.Before(t2)
		}
	default:
		// no-op; leave existing order
		return
	}

	if order == "desc" {
		sort.SliceStable(slice, func(i, j int) bool { return !less(i, j) })
	} else {
		sort.SliceStable(slice, less)
	}
}
