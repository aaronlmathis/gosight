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
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
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
func (t *AgentTracker) UpdateAgent(meta model.Meta) {
	t.mu.Lock()
	defer t.mu.Unlock()

	a, ok := t.agents[meta.Hostname]
	if !ok {
		a = &model.AgentStatus{
			Hostname: meta.Hostname,
			IP:       meta.PrivateIP,
			Labels:   meta.Tags,
		}
		t.agents[meta.Hostname] = a
	}

	a.LastSeen = time.Now()
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
			Labels:   a.Labels,
			Status:   status,
			Since:    elapsed.Truncate(time.Second).String(),
		})
	}
	return list
}

func deriveStatus(lastSeen time.Time) string {
	elapsed := time.Since(lastSeen)

	switch {
	case elapsed < 10*time.Second:
		return "Online"
	case elapsed < 60*time.Second:
		return "Idle"
	default:
		return "Offline"
	}
}
