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
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/events"
	"github.com/aaronlmathis/gosight/server/internal/store/datastore"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/google/uuid"
)

type AgentTracker struct {
	mu        sync.RWMutex
	agents    map[string]*model.Agent
	emitter   *events.Emitter
	ctx       context.Context
	dataStore datastore.DataStore
}

// Create a new tracker
func NewAgentTracker(ctx context.Context, emitter *events.Emitter, dataStore datastore.DataStore) *AgentTracker {
	return &AgentTracker{
		agents:    make(map[string]*model.Agent),
		emitter:   emitter,
		ctx:       ctx,
		dataStore: dataStore,
	}
}

// Updates in memory store of Agent details
func (t *AgentTracker) UpdateAgent(meta *model.Meta) {
	//utils.Debug("Entering UpdateAgent")
	if meta.Hostname == "" || meta.ContainerID != "" {
		return
	}
	utils.Debug("📥 UpdateAgent: %s (agent_id=%s) at %s", meta.Hostname, meta.AgentID, time.Now().Format(time.RFC3339))
	t.mu.Lock()
	defer t.mu.Unlock()

	agent, exists := t.agents[meta.AgentID]
	if !exists {
		if t.dataStore != nil {
			existing, err := t.dataStore.GetAgentByID(t.ctx, meta.AgentID)
			if err == nil && existing != nil {
				// Already existed in DB → no "registered" emit
				agent = existing
			} else {
				// Truly new agent → emit "registered"
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
				t.emitter.Emit(t.ctx, model.EventEntry{
					Level:      "info",
					Category:   "system",
					Message:    fmt.Sprintf("Agent %s registered", agent.Hostname),
					Source:     "agent.lifecycle",
					Scope:      "endpoint",
					Target:     agent.AgentID,
					EndpointID: agent.EndpointID,
					Meta:       BuildAgentEventMeta(agent),
				})
			}
			t.agents[meta.AgentID] = agent
		} else {
			// fallback path if datastore isn't initialized
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
		}
	}

	if startRaw, ok := meta.Tags["agent_start_time"]; ok {
		if startUnix, err := strconv.ParseInt(startRaw, 10, 64); err == nil {
			if agent.StartTime.IsZero() {
				agent.StartTime = time.Unix(startUnix, 0)
				//utils.Debug("meta.Tags[agent_start_time]: %s", meta.Tags["agent_start_time"])
			}
			agent.UptimeSeconds = time.Since(agent.StartTime).Seconds()
		} else {
			utils.Warn("Invalid agent_start_time tag: %s", startRaw)
		}
	}
	agent.LastSeen = time.Now()
}

// GetAll returns all agents from  MEMORY (only online agents)
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

func (t *AgentTracker) GetAgentMap() map[string]model.Agent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[string]model.Agent)

	for id, a := range t.agents {
		elapsed := time.Since(a.LastSeen)

		status := "Offline"
		if elapsed < 10*time.Second {
			status = "Online"
		} else if elapsed < 60*time.Second {
			status = "Idle"
		}

		uptime := 0.0
		if !a.StartTime.IsZero() {
			uptime = time.Since(a.StartTime).Seconds()
		}

		result[id] = model.Agent{
			AgentID:       a.AgentID,
			HostID:        a.HostID,
			Hostname:      a.Hostname,
			IP:            a.IP,
			OS:            a.OS,
			Arch:          a.Arch,
			Version:       a.Version,
			Labels:        a.Labels,
			EndpointID:    a.EndpointID,
			LastSeen:      a.LastSeen,
			Status:        status,
			Since:         elapsed.Truncate(time.Second).String(),
			UptimeSeconds: uptime,
		}
	}

	return result
}

// Syncs Agents from inmemory to persistant storage
func (t *AgentTracker) SyncToStore(ctx context.Context, store datastore.DataStore) {
	utils.Debug("Sync to Store called")
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, agent := range t.agents {
		elapsed := time.Since(agent.LastSeen)

		status := "Offline"
		if elapsed < 10*time.Second {
			status = "Online"
		} else if elapsed < 60*time.Second {
			status = "Idle"
		}

		if agent.Status != status {
			agent.Status = status
			agent.Updated = true
		}

		if !agent.Updated {
			continue
		}
		utils.Debug("Syncing agent %s to store: %s", agent.Hostname, status)
		err := store.UpsertAgent(ctx, agent)
		if err != nil {
			utils.Error("Agent sync failed for %s: %v", agent.Hostname, err)
			continue
		}

		agent.Updated = false
	}
}

func (t *AgentTracker) CheckAgentStatusesAndEmitEvents() {
	utils.Debug("🔁 Checking agent lifecycle statuses...")
	now := time.Now()
	storedAgents, err := t.dataStore.ListAgents(t.ctx)
	utils.Debug("📦 Loaded %d agents from store", len(storedAgents))
	if err != nil {
		utils.Error("Agent lifecycle check: failed to list agents: %v", err)
		return
	}

	for _, agent := range storedAgents {
		isLive := t.IsLive(agent.AgentID)

		utils.Debug("🧠 Agent %s | live=%v | status=%s | age=%v", agent.Hostname, isLive, agent.Status, time.Since(agent.LastSeen))

		//  Agent is missing from in-memory tracker and was Online → emit Offline
		if !isLive && agent.Status != "Offline" && time.Since(agent.LastSeen) > 2*time.Minute {
			utils.Debug("⚠️ Agent %s marked offline (last seen %s)", agent.Hostname, agent.LastSeen.Format(time.RFC3339))
			agent.Status = "Offline"
			t.emitter.Emit(t.ctx, model.EventEntry{
				ID:         uuid.NewString(),
				Timestamp:  now,
				Level:      "warning",
				Type:       "event",
				Category:   "system",
				Message:    fmt.Sprintf("Agent %s went offline", agent.Hostname),
				Source:     "agent.lifecycle",
				Scope:      "endpoint",
				Target:     agent.AgentID,
				EndpointID: agent.EndpointID,
				Meta:       BuildAgentEventMeta(agent),
			})
			agent.Updated = true
		}

		//  Agent is back in in-memory but was marked Offline → emit Back Online
		if isLive && agent.Status != "Online" {
			utils.Debug("✅ Agent %s is back online", agent.Hostname)
			agent.Status = "Online"
			t.emitter.Emit(t.ctx, model.EventEntry{
				ID:         uuid.NewString(),
				Timestamp:  now,
				Level:      "info",
				Type:       "event",
				Category:   "system",
				Message:    fmt.Sprintf("Agent %s is back online", agent.Hostname),
				Source:     "agent.lifecycle",
				Scope:      "endpoint",
				Target:     agent.AgentID,
				EndpointID: agent.EndpointID,
				Meta:       BuildAgentEventMeta(agent),
			})
			agent.Updated = true
		}
	}
}
func (t *AgentTracker) IsLive(agentID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if a, ok := t.agents[agentID]; ok {
		return time.Since(a.LastSeen) <= 10*time.Second
	}
	return false
}
