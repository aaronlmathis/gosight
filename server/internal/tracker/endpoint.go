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

package tracker

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

type EndpointTracker struct {
	mu         sync.RWMutex
	agents     map[string]*model.Agent
	containers map[string]*model.Container
	emitter    *events.Emitter
	ctx        context.Context
	dataStore  datastore.DataStore
}

// NewEndpointTracker creates a new EndpointTracker instance.
// It initializes the in-memory store for agent and container details and sets up the event emitter.
// The EndpointTracker is responsible for tracking agent and container heartbeats and updating their status.
// It also interacts with the datastore to persist agent and container information.
func NewEndpointTracker(ctx context.Context, emitter *events.Emitter, dataStore datastore.DataStore) *EndpointTracker {
	return &EndpointTracker{
		agents:     make(map[string]*model.Agent),
		containers: make(map[string]*model.Container),
		emitter:    emitter,
		ctx:        ctx,
		dataStore:  dataStore,
	}
}

// Updates in memory store of Agent details
func (t *EndpointTracker) UpdateAgent(meta *model.Meta) {
	//utils.Debug("Entering UpdateAgent")
	if meta.Hostname == "" || meta.ContainerID != "" {
		return
	}
	//utils.Debug("UpdateAgent: %s (agent_id=%s) at %s", meta.Hostname, meta.AgentID, time.Now().Format(time.RFC3339))
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

// UpdateContainer updates the container details in the in-memory store.
func (t *EndpointTracker) UpdateContainer(meta *model.Meta) {
	if meta.ContainerID == "" {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	c, exists := t.containers[meta.ContainerID]
	now := time.Now()

	newStatus := meta.Tags["status"]
	newStart := meta.Tags["agent_start_time"]

	if !exists {
		// New container observed
		c = &model.Container{
			ContainerID: meta.ContainerID,
			Name:        meta.ContainerName,
			ImageName:   meta.ContainerImageName,
			ImageID:     meta.ContainerImageID,
			Runtime:     meta.Tags["job"],
			Status:      newStatus,
			HostID:      meta.HostID,
			EndpointID:  meta.EndpointID,
			LastSeen:    now,
			Labels:      meta.Tags,
			Updated:     true,
		}
		t.containers[meta.ContainerID] = c

		t.emitter.Emit(t.ctx, model.EventEntry{
			ID:         uuid.NewString(),
			Timestamp:  now,
			Type:       "event",
			Level:      "info",
			Category:   "container",
			Message:    fmt.Sprintf("Container %s started", c.Name),
			Source:     "container.lifecycle",
			Scope:      "container",
			Target:     c.ContainerID,
			EndpointID: c.EndpointID,
			Meta:       BuildContainerEventMeta(c),
		})
		return
	}

	// Detect restart
	prevStart := c.Labels["agent_start_time"]
	if newStart != "" && prevStart != "" && newStart != prevStart {
		t.emitter.Emit(t.ctx, model.EventEntry{
			ID:         uuid.NewString(),
			Timestamp:  now,
			Type:       "event",
			Level:      "info",
			Category:   "container",
			Message:    fmt.Sprintf("Container %s restarted", c.Name),
			Source:     "container.lifecycle",
			Scope:      "container",
			Target:     c.ContainerID,
			EndpointID: c.EndpointID,
			Meta:       BuildContainerEventMeta(c),
		})
	}

	// Auto-resolve if it was marked inactive
	if c.Status == "inactive" && newStatus != "inactive" {
		t.emitter.Emit(t.ctx, model.EventEntry{
			ID:         uuid.NewString(),
			Timestamp:  now,
			Type:       "event",
			Level:      "info",
			Category:   "container",
			Message:    fmt.Sprintf("Container %s is back online", c.Name),
			Source:     "container.lifecycle",
			Scope:      "container",
			Target:     c.ContainerID,
			EndpointID: c.EndpointID,
			Meta:       BuildContainerEventMeta(c),
		})
	}

	// Update current state
	c.LastSeen = now
	c.Status = newStatus
	c.Labels = meta.Tags
	c.EndpointID = meta.EndpointID
	c.HostID = meta.HostID
	c.Updated = true
}

// GetAgents returns all agents from  MEMORY (only online agents)
func (t *EndpointTracker) GetAgents() []model.Agent {
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

// GetAgentMap returns a map of agent details, including their status and uptime.
// The map is keyed by agent ID and contains information such as hostname, IP address,
// OS, architecture, version, labels, endpoint ID, last seen time, status, and uptime in seconds.
func (t *EndpointTracker) GetAgentMap() map[string]model.Agent {
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

// SyncToStore writes all updated agents and containers to persistent storage.
func (t *EndpointTracker) SyncToStore(ctx context.Context, store datastore.DataStore) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// --- Sync Agents ---
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

		utils.Debug("Syncing agent %s to store: %s", agent.Hostname, agent.Status)
		if err := t.dataStore.UpsertAgent(ctx, agent); err != nil {
			utils.Error("❌ Agent sync failed for %s: %v", agent.Hostname, err)
			continue
		}
		agent.Updated = false
	}

	// --- Sync Containers ---
	for _, ctr := range t.containers {
		if !ctr.Updated {
			continue
		}

		utils.Debug("Syncing container %s to t.dataStore: %s", ctr.Name, ctr.Status)
		if err := t.dataStore.UpsertContainer(ctx, ctr); err != nil {
			utils.Error("❌ Container sync failed for %s: %v", ctr.Name, err)
			continue
		}
		ctr.Updated = false
	}
}

// CheckAgentStatusesAndEmitEvents checks the status of agents and emits events
// for agents that have gone offline or come back online.
// It updates the agent status in the datastore and emits lifecycle events.
// This function is called periodically to ensure the agent statuses are up to date.
func (t *EndpointTracker) CheckAgentStatusesAndEmitEvents() {
	//utils.Debug("Checking agent lifecycle statuses...")
	now := time.Now()
	storedAgents, err := t.dataStore.ListAgents(t.ctx)
	//utils.Debug("Loaded %d agents from store", len(storedAgents))
	if err != nil {
		utils.Error("Agent lifecycle check: failed to list agents: %v", err)
		return
	}

	for _, agent := range storedAgents {
		isLive := t.IsLive(agent.AgentID)

		//  Agent is missing from in-memory tracker and was Online → emit Offline
		if !isLive && agent.Status != "Offline" && time.Since(agent.LastSeen) > 2*time.Minute {
			//utils.Debug("Agent %s marked offline (last seen %s)", agent.Hostname, agent.LastSeen.Format(time.RFC3339))
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
			//utils.Debug(" Agent %s is back online", agent.Hostname)
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
		if agent.Updated {
			//utils.Debug("Writing updated agent status for %s", agent.Hostname)
			err := t.dataStore.UpsertAgent(t.ctx, agent)
			if err != nil {
				utils.Error("❌ Failed to upsert agent %s: %v", agent.Hostname, err)
			}
			agent.Updated = false
		}
	}
}

// IsLive checks if an agent is live based on its last seen time.
// It returns true if the agent was seen within the last 10 seconds, otherwise false.
func (t *EndpointTracker) IsLive(agentID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if a, ok := t.agents[agentID]; ok {
		return time.Since(a.LastSeen) <= 10*time.Second
	}
	return false
}

// CheckContainerStatusesAndEmit checks the status of containers and emits events
// for containers that have gone offline.
// It updates the container status in the datastore and emits lifecycle events.
// This function is called periodically to ensure the container statuses are up to date.
func (t *EndpointTracker) CheckContainerStatusesAndEmit() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	// Step 1: Pull known containers from store
	storedContainers, err := t.dataStore.ListContainers(t.ctx)
	if err != nil {
		utils.Error("Failed to list containers for lifecycle check: %v", err)
		return
	}

	for _, stored := range storedContainers {
		c, exists := t.containers[stored.ContainerID]

		if !exists || time.Since(c.LastSeen) > 2*time.Minute {
			// If it’s not in memory OR stale
			if stored.Status != "inactive" {
				stored.Status = "inactive"
				stored.Updated = true

				t.emitter.Emit(t.ctx, model.EventEntry{
					ID:         uuid.NewString(),
					Timestamp:  now,
					Type:       "event",
					Level:      "warning",
					Category:   "container",
					Message:    fmt.Sprintf("Container %s is no longer reporting (last known: %s)", stored.Name, stored.Status),
					Source:     "container.lifecycle",
					Scope:      "container",
					Target:     stored.ContainerID,
					EndpointID: stored.EndpointID,
					Meta:       BuildContainerEventMeta(stored),
				})
			}
		}
	}
}
