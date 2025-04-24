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

	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// RenderAgentsPage serves the agents.html template
func (s *HttpServer) HandleAgentsPage(w http.ResponseWriter, r *http.Request, templateDir, env string) {
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

/*
// HandleAgentsAPI returns a JSON list of active agents

	func (s *HttpServer) HandleAgentsAPI(w http.ResponseWriter, r *http.Request) {
		agents := s.Sys.Agents.GetAgents()

		sort.SliceStable(agents, func(i, j int) bool {
			return agents[i].Hostname < agents[j].Hostname
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(agents)
	}
*/
func (s *HttpServer) HandleAgentsAPI(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	userID, _ := contextutil.GetUserID(r.Context())
	roles, _ := contextutil.GetUserRoles(r.Context())
	perms, _ := contextutil.GetUserPermissions(r.Context())

	utils.Debug("🧑‍💻 API /agents called by: %s | Roles: %v | Perms: %v", userID, roles, perms)
	storedAgents, err := s.Sys.Stores.Data.ListAgents(ctx)

	if err != nil {
		http.Error(w, "failed to load agents", http.StatusInternalServerError)
		return
	}
	if storedAgents == nil {
		storedAgents = []*model.Agent{}
	}
	// 2. Get current in-memory live agents
	liveMap := s.Sys.Agents.GetAgentMap()

	// 3. Merge: overwrite stored fields with live status if found
	for i := range storedAgents {
		if live, ok := liveMap[storedAgents[i].AgentID]; ok {
			storedAgents[i].Status = live.Status
			storedAgents[i].Since = live.Since
			storedAgents[i].UptimeSeconds = live.UptimeSeconds
			storedAgents[i].LastSeen = live.LastSeen
		} else {
			storedAgents[i].Status = "Offline"
			storedAgents[i].Since = "—"
		}
	}

	// 4. Sort by hostname
	sort.SliceStable(storedAgents, func(i, j int) bool {
		return storedAgents[i].Hostname < storedAgents[j].Hostname
	})

	// 5. Respond
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(storedAgents)

}
