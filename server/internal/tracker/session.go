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
	"time"

	"github.com/aaronlmathis/gosight/shared/proto"
)

type LiveAgentSession struct {
	Stream        proto.StreamService_StreamServer
	ConnectedAt   time.Time
	LastHeartbeat time.Time
}

// RegisterAgentSession registers a live connected agent
func (t *EndpointTracker) RegisterAgentSession(agentID string, client proto.StreamService_StreamServer) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sessions[agentID] = &LiveAgentSession{
		Stream:        client,
		ConnectedAt:   time.Now(),
		LastHeartbeat: time.Now(),
	}
}

// GetAgentSession retrieves a live agent session
func (t *EndpointTracker) GetAgentSession(agentID string) (*LiveAgentSession, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	sess, ok := t.sessions[agentID]
	return sess, ok
}

// RemoveAgentSession cleans up after disconnect
func (t *EndpointTracker) RemoveAgentSession(agentID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.sessions, agentID)
}
