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

// gosight/server/internal/dispatcher/dispatcher.go

// Package dispatcher provides functionality to manage and dispatch
package dispatcher

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type Dispatcher struct {
	Routes map[string]model.ActionRoute
}

func NewDispatcher(routeMap map[string]model.ActionRoute) *Dispatcher {
	return &Dispatcher{Routes: routeMap}
}

// Dispatch processes an event against all routes and triggers matching actions.
func (d *Dispatcher) Dispatch(ctx context.Context, event model.EventEntry) {
	for _, route := range d.Routes {
		if !matchRoute(route.Match, event) {
			continue
		}
		utils.Debug("Dispatching event:" + event.Message)
		for _, action := range route.Actions {
			go d.ExecuteAction(ctx, action, event)
		}
	}
}

// TriggerActionByID looks up a route by ID and executes its actions.
func (d *Dispatcher) TriggerActionByID(ctx context.Context, actionID string, event model.EventEntry) {
	route, ok := d.Routes[actionID]
	if !ok {
		utils.Warn("🚫 No route found for action ID: %s", actionID)
		return
	}

	for _, action := range route.Actions {
		d.ExecuteAction(ctx, action, event)
	}
}

// matchRoute checks if the event matches the route's filter criteria.
// It compares the event's level, rule ID, and tags against the filter.
// If all criteria match, it returns true; otherwise, false.
func matchRoute(f model.MatchFilter, e model.EventEntry) bool {
	if f.Level != "" && f.Level != e.Level {
		return false
	}
	if f.RuleID != "" && f.RuleID != e.Meta["rule_id"] {
		return false
	}
	for k, v := range f.Tags {
		if e.Meta[k] != v {
			return false
		}
	}
	return true
}

// executeAction executes the action specified in the route.
// It determines the action type (webhook or script) and calls the appropriate function.
func (d *Dispatcher) ExecuteAction(ctx context.Context, a model.ActionSpec, e model.EventEntry) {
	switch strings.ToLower(a.Type) {
	case "webhook":
		executeWebhook(a, e)
	case "script":
		executeScript(a, e)
	default:
		// unknown type
	}
}

// executeWebhook sends a POST request to the specified URL with the event data as JSON.
// It sets the content type to application/json and includes any additional headers specified in the action.
func executeWebhook(a model.ActionSpec, e model.EventEntry) {
	utils.Debug("🔥 Webhook triggered for: %s", a.URL)
	payload, _ := json.Marshal(e)
	req, _ := http.NewRequest("POST", a.URL, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	for k, v := range a.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	client.Do(req) // ignore error here, but should log
}

// executeScript runs a script with the event data as input.
// It uses the command and arguments specified in the action's Command and Args fields.
// The event data is passed to the script via standard input.
// The script is expected to handle the input and perform the necessary actions.
func executeScript(a model.ActionSpec, e model.EventEntry) {
	payload, _ := json.Marshal(e)
	cmd := exec.Command(a.Command, a.Args...)
	cmd.Stdin = bytes.NewReader(payload)
	cmd.Run() // ignore error here, but should log
}
