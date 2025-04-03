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

// Basic Handler for http server
// server/internal/http/handler.go
package httpserver

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

var demoAgents = []model.AgentStatus{
	{Name: "agent-01", Status: "online", LastSeen: "3s ago", IP: "192.168.1.101", Zone: "DC-1", CPU: 22.5},
	{Name: "agent-02", Status: "idle", LastSeen: "45s ago", IP: "192.168.1.102", Zone: "DC-2", CPU: 9.1},
	{Name: "agent-03", Status: "offline", LastSeen: "10m ago", IP: "192.168.1.103", Zone: "Edge", CPU: 0.0},
}

var tracker *store.AgentTracker

func InitHandlers(t *store.AgentTracker) {
	tracker = t
}

func handleIndex(w http.ResponseWriter, r *http.Request, templateDir, env string) {
	tmplPath := filepath.Join(templateDir, "index.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		utils.Error("Template parse error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title": "GoSight",
		"Env":   env,
	}

	tmpl.Execute(w, data)
}

func handleAgentsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(demoAgents); err != nil {
		utils.Error("Failed to encode agent list: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
