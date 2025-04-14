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

// Store agent details/heartbeats
// server/internal/store/agent_tracker.go
package store

import (
	"context"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/store/datastore"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type AgentTracker struct {
	mu     sync.RWMutex
	agents map[string]*model.AgentStatus
}

// Create a new tracker
func NewAgentTracker() *AgentTracker {
	return &AgentTracker{
		agents: make(map[string]*model.AgentStatus),
	}
}

// Updates in memory store of Agent details
func (t *AgentTracker) UpdateAgent(meta model.Meta) {

	if meta.Hostname == "" {
		// Don't track nameless agents
		utils.Warn("Skipping UpdateAgent: meta.Hostname is empty")
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	agent, exists := t.agents[meta.Hostname]
	if !exists {
		agent = &model.AgentStatus{
			Hostname: meta.Hostname,
			IP:       meta.IPAddress,
			OS:       meta.OS,
			Arch:     meta.Architecture,
			Version:  meta.AgentVersion,
			Labels:   meta.Tags,
			Updated:  true,
		}
		t.agents[meta.Hostname] = agent
	} else {
		// Update any fields that may change
		agent.IP = meta.IPAddress
		agent.OS = meta.OS
		agent.Arch = meta.Architecture
		agent.Version = meta.AgentVersion
		agent.Labels = meta.Tags
		agent.Updated = true
	}

	agent.LastSeen = time.Now()
}

func (t *AgentTracker) GetAgents() []model.AgentStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var list []model.AgentStatus
	for _, a := range t.agents {
		elapsed := time.Since(a.LastSeen)

		// Derive status
		status := "Offline"
		if elapsed < 10*time.Second {
			status = "Online"
		} else if elapsed < 60*time.Second {
			status = "Idle"
		}

		list = append(list, model.AgentStatus{
			Hostname: a.Hostname,
			IP:       a.IP,
			OS:       a.OS,
			Labels:   a.Labels,
			Status:   status,
			Since:    elapsed.Truncate(time.Second).String(),
		})
	}
	return list
}

// Syncs Agents from inmemory to persistant storage
func (t *AgentTracker) SyncToStore(ctx context.Context, store datastore.DataStore) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, agent := range t.agents {
		if !agent.Updated {
			continue
		}

		err := store.UpsertAgent(ctx, agent)
		if err != nil {
			utils.Error("❌ Agent sync failed for %s: %v", agent.Hostname, err)
			continue
		}

		agent.Updated = false
	}
}
