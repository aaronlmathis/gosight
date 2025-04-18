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
package agenttracker

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
	agents map[string]*model.Agent
}

// Create a new tracker
func NewAgentTracker() *AgentTracker {
	return &AgentTracker{
		agents: make(map[string]*model.Agent),
	}
}

// Updates in memory store of Agent details
func (t *AgentTracker) UpdateAgent(meta *model.Meta) {
	if meta.Hostname == "" || meta.ContainerID != "" {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	agent, exists := t.agents[meta.AgentID]
	if !exists {
		agent = &model.Agent{
			AgentID:    meta.AgentID,
			HostID:     meta.HostID,
			Hostname:   meta.Hostname,
			IP:         meta.IPAddress,
			OS:         meta.OS,
			Arch:       meta.Architecture,
			Version:    meta.AgentVersion,
			Labels:     meta.Tags,
			EndpointID: meta.EndpointID,
			Updated:    true,
		}
		t.agents[meta.AgentID] = agent
	} else {
		// Update mutable fields
		agent.IP = meta.IPAddress
		agent.OS = meta.OS
		agent.Arch = meta.Architecture
		agent.Version = meta.AgentVersion
		agent.Labels = meta.Tags
		agent.EndpointID = meta.EndpointID
		agent.Updated = true
	}

	agent.LastSeen = time.Now()
}

func (t *AgentTracker) GetAgents() []model.Agent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var list []model.Agent
	now := time.Now()
	for _, a := range t.agents {
		elapsed := now.Sub(a.LastSeen)

		// Derive status
		status := "Offline"
		if elapsed < 10*time.Second {
			status = "Online"
		} else if elapsed < 60*time.Second {
			status = "Idle"
		}

		list = append(list, model.Agent{
			AgentID:    a.AgentID,
			HostID:     a.HostID,
			Hostname:   a.Hostname,
			IP:         a.IP,
			OS:         a.OS,
			Arch:       a.Arch,
			Version:    a.Version,
			Labels:     a.Labels,
			EndpointID: a.EndpointID,
			Status:     status,
			Since:      elapsed.Truncate(time.Second).String(),
		})
	}
	return list
}

func (t *AgentTracker) GetAgentMap() map[string]*model.Agent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a shallow copy so we don’t expose internal map
	out := make(map[string]*model.Agent, len(t.agents))
	for id, agent := range t.agents {
		out[id] = agent
	}
	return out
}

// Syncs Agents from inmemory to persistant storage
func (t *AgentTracker) SyncToStore(ctx context.Context, store datastore.DataStore) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for hostname, agent := range t.agents {
		utils.Debug("Syncing agent: %s | EndpointID: %s | IP: %s", hostname, agent.AgentID, agent.IP)
	}
	for _, agent := range t.agents {
		if !agent.Updated {
			continue
		}

		err := store.UpsertAgent(ctx, agent)
		if err != nil {
			utils.Error("Agent sync failed for %s: %v", agent.Hostname, err)
			continue
		}

		agent.Updated = false
	}
}
