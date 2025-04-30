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

package tracker

import (
	"context"
	"fmt"
	"math"
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
	mu           sync.RWMutex
	agents       map[string]*model.Agent
	containers   map[string]*model.Container
	emitter      *events.Emitter
	ctx          context.Context
	dataStore    datastore.DataStore
	sessions     map[string]*LiveAgentSession
	commandQueue map[string]*CommandQueue
}

func NewEndpointTracker(ctx context.Context, emitter *events.Emitter, dataStore datastore.DataStore) *EndpointTracker {
	agents := make(map[string]*model.Agent)
	containers := make(map[string]*model.Container)

	storedAgents, err := dataStore.ListAgents(ctx)
	if err != nil {
		utils.Error("Failed to load agents from datastore: %v", err)
	}

	storedContainers, err := dataStore.ListContainers(ctx)
	if err != nil {
		utils.Error("Failed to load containers from datastore: %v", err)
	}

	for _, agent := range storedAgents {
		agents[agent.AgentID] = agent
	}
	for _, container := range storedContainers {
		containers[container.ContainerID] = container
	}

	return &EndpointTracker{
		agents:       agents,
		containers:   containers,
		emitter:      emitter,
		ctx:          ctx,
		sessions:     make(map[string]*LiveAgentSession),
		commandQueue: make(map[string]*CommandQueue),
		dataStore:    dataStore,
	}
}

func (t *EndpointTracker) UpdateAgent(meta *model.Meta) {
	if meta.Hostname == "" || meta.ContainerID != "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	agent, exists := t.agents[meta.AgentID]
	if !exists {
		if t.dataStore != nil {
			existing, err := t.dataStore.GetAgentByID(t.ctx, meta.AgentID)
			if err == nil && existing != nil {
				agent = existing
			} else {
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
			}
			agent.UptimeSeconds = time.Since(agent.StartTime).Seconds()
		} else {
			utils.Warn("Invalid agent_start_time tag: %s", startRaw)
		}
	}
	agent.LastSeen = time.Now()
}

func (t *EndpointTracker) UpdateContainer(meta *model.Meta) {
	if meta.ContainerID == "" {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	c, exists := t.containers[meta.ContainerID]
	now := time.Now()

	newStatus := NormalizeContainerStatus(meta.Tags["status"])
	newStart := meta.Tags["agent_start_time"]

	if !exists {
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

	prevStart := c.Labels["agent_start_time"]
	if prevStart != "" && newStart != "" && newStart != prevStart {
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

	// Detect container engine status change
	if c.Status != newStatus {
		utils.Debug("Container %s engine status changed from %s to %s", c.Name, c.Status, newStatus)

		level := "info"
		if newStatus == "Exited" || newStatus == "Stopped" {
			level = "warning"
		}

		t.emitter.Emit(t.ctx, model.EventEntry{
			ID:         uuid.NewString(),
			Timestamp:  now,
			Type:       "event",
			Level:      level,
			Category:   "container",
			Message:    fmt.Sprintf("Container %s changed engine status to %s", c.Name, newStatus),
			Source:     "container.lifecycle",
			Scope:      "container",
			Target:     c.ContainerID,
			EndpointID: c.EndpointID,
			Meta:       BuildContainerEventMeta(c),
		})

		c.Status = newStatus // VERY IMPORTANT
		c.Updated = true
	}

	// Detect return from "inactive" separately (optional logic if needed)
	if c.Status == "Inactive" && newStatus != "Inactive" {
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

	// Always refresh timestamps and labels
	c.LastSeen = now
	c.Labels = meta.Tags
	c.EndpointID = meta.EndpointID
	c.HostID = meta.HostID
	c.Updated = true
}

func (t *EndpointTracker) GetAgents() []model.Agent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var list []model.Agent
	now := time.Now()
	for _, a := range t.agents {
		elapsed := now.Sub(a.LastSeen)

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
			uptime = math.Round(time.Since(a.StartTime).Seconds())
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

func (t *EndpointTracker) SyncToStore(ctx context.Context, store datastore.DataStore) {
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

		utils.Debug("Syncing agent %s to store: %s", agent.Hostname, agent.Status)
		if err := t.dataStore.UpsertAgent(ctx, agent); err != nil {
			utils.Error("Agent sync failed for %s: %v", agent.Hostname, err)
			continue
		}
		agent.Updated = false
	}

	for _, ctr := range t.containers {
		if !ctr.Updated {
			continue
		}

		utils.Debug("Syncing container %s to store: %s (Heartbeat: %s)", ctr.Name, ctr.Status, ctr.Heartbeat)
		if err := t.dataStore.UpsertContainer(ctx, ctr); err != nil {
			utils.Error("Container sync failed for %s: %v", ctr.Name, err)
			continue
		}
		ctr.Updated = false
	}
}

func (t *EndpointTracker) CheckAgentStatusesAndEmitEvents() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	for _, agent := range t.agents {
		elapsed := now.Sub(agent.LastSeen)

		status := "Offline"
		if elapsed < 10*time.Second {
			status = "Online"
		} else if elapsed < 60*time.Second {
			status = "Idle"
		}

		if agent.Status != status {
			utils.Debug("Updating agent %s status from %s to %s", agent.Hostname, agent.Status, status)
			agent.Status = status
			agent.Updated = true

			level := "info"
			if status == "Offline" {
				level = "warning"
			}
			t.emitter.Emit(t.ctx, model.EventEntry{
				ID:         uuid.NewString(),
				Timestamp:  now,
				Type:       "event",
				Level:      level,
				Category:   "system",
				Message:    fmt.Sprintf("Agent %s changed status to %s", agent.Hostname, status),
				Source:     "agent.lifecycle",
				Scope:      "endpoint",
				Target:     agent.AgentID,
				EndpointID: agent.EndpointID,
				Meta:       BuildAgentEventMeta(agent),
			})
		}
	}
}

func (t *EndpointTracker) CheckContainerStatusesAndEmit() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	for _, container := range t.containers {
		elapsed := now.Sub(container.LastSeen)

		newHeartbeat := "Offline"
		if elapsed < 10*time.Second {
			newHeartbeat = "Online"
		} else if elapsed < 60*time.Second {
			newHeartbeat = "Idle"
		}

		if container.Heartbeat != newHeartbeat {
			utils.Debug("Updating container %s status from %s to %s", container.Name, container.Heartbeat, newHeartbeat)
			container.Heartbeat = newHeartbeat
			container.Updated = true

			level := "info"
			if newHeartbeat == "Offline" {
				level = "warning"
			}

			t.emitter.Emit(t.ctx, model.EventEntry{
				ID:         uuid.NewString(),
				Timestamp:  now,
				Type:       "event",
				Level:      level,
				Category:   "container",
				Message:    fmt.Sprintf("Container %s changed status to %s", container.Name, container.Heartbeat),
				Source:     "container.lifecycle",
				Scope:      "container",
				Target:     container.ContainerID,
				EndpointID: container.EndpointID,
				Meta:       BuildContainerEventMeta(container),
			})
		}
	}
}

func (t *EndpointTracker) IsAgentLive(agentID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if a, ok := t.agents[agentID]; ok {
		return time.Since(a.LastSeen) <= 10*time.Second
	}
	return false
}

func (t *EndpointTracker) IsContainerLive(endpointID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if a, ok := t.containers[endpointID]; ok {
		return time.Since(a.LastSeen) <= 10*time.Second
	}
	return false
}

func (t *EndpointTracker) ListEndpoints() []model.Endpoint {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var endpoints []model.Endpoint
	for _, agent := range t.agents {
		endpoints = append(endpoints, model.Endpoint{
			EndpointID: agent.EndpointID,
			HostID:     agent.HostID,
			Hostname:   agent.Hostname,
			Labels:     agent.Labels,
			LastSeen:   agent.LastSeen,
			Status:     agent.Status,
		})
	}

	for _, container := range t.containers {
		endpoints = append(endpoints, model.Endpoint{
			EndpointID:    container.EndpointID,
			HostID:        container.HostID,
			ContainerName: container.Name,
			Labels:        container.Labels,
			LastSeen:      container.LastSeen,
			Status:        container.Heartbeat,
		})
	}

	return endpoints
}

func (t *EndpointTracker) GetEndpointIdByAgentId(agentID string) (string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if agent, ok := t.agents[agentID]; ok {
		return agent.EndpointID, true
	}
	return "", false
}
