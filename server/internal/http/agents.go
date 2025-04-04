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

// server/internal/http/agents.go

package httpserver

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/aaronlmathis/gosight/shared/utils"
)

// RenderAgentsPage serves the agents.html template
func RenderAgentsPage(w http.ResponseWriter, r *http.Request, templateDir, env string) {
	tmplPath := filepath.Join(templateDir, "agents.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		utils.Error("Template parse error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title": "Agents - GoSight",
		"Env":   env,
	}

	_ = tmpl.Execute(w, data)
}

// HandleAgentsAPI returns a JSON list of active agents
func HandleAgentsAPI(w http.ResponseWriter, r *http.Request) {
	agents := tracker.GetAgents()

	sort.SliceStable(agents, func(i, j int) bool {
		return agents[i].Hostname < agents[j].Hostname
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(agents)
}
