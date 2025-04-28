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

	"github.com/aaronlmathis/gosight/shared/model"
)

func (s *HttpServer) HandleAlertsAPI(w http.ResponseWriter, r *http.Request) {

	alerts, err := s.Sys.Stores.Alerts.ListAlerts(r.Context())
	if err != nil {
		http.Error(w, "failed to load alerts", http.StatusInternalServerError)
		return
	}

	if alerts == nil {
		alerts = []model.AlertInstance{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		http.Error(w, "failed to encode alerts", http.StatusInternalServerError)
		return
	}
}

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
