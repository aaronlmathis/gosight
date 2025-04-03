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
	agents map[string]*agentState
}

type agentState struct {
	status   model.AgentStatus
	lastSeen time.Time
}

// Create a new tracker
func NewAgentTracker() *AgentTracker {
	return &AgentTracker{
		agents: make(map[string]*agentState),
	}
}

func (t *AgentTracker) UpdateAgent(id string, status model.AgentStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()

	status.LastSeen = time.Now().Format("15:04:05") // optional
	t.agents[id] = &agentState{
		status:   status,
		lastSeen: time.Now(),
	}
}
