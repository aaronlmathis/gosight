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
	"github.com/aaronlmathis/gosight/shared/utils"
)

type LiveAgentSession struct {
	Stream        proto.StreamService_StreamServer
	ConnectedAt   time.Time
	LastHeartbeat time.Time
	SendQueue     chan *proto.StreamResponse
}

// RegisterAgentSession registers a live connected agent
func (t *EndpointTracker) RegisterAgentSession(agentID string, client proto.StreamService_StreamServer) {
	t.mu.Lock()
	session := &LiveAgentSession{
		Stream:        client,
		ConnectedAt:   time.Now(),
		LastHeartbeat: time.Now(),
		SendQueue:     make(chan *proto.StreamResponse, 10),
	}
	t.sessions[agentID] = session
	t.mu.Unlock()

	//utils.Info("Registered agent session: %s", agentID)

	// Start dedicated send loop
	go func() {
		for resp := range session.SendQueue {
			err := session.Stream.Send(resp)
			if err != nil {
				utils.Warn("Failed to send StreamResponse to agent %s: %v", agentID, err)
				// optional: remove the session and cleanup if the stream is broken
				t.mu.Lock()
				delete(t.sessions, agentID)
				t.mu.Unlock()
				return
			}
			//utils.Debug("StreamResponse sent to %s", agentID)
		}
	}()
}

// GetAgentSession retrieves a live agent session
func (t *EndpointTracker) GetAgentSession(agentID string) (*LiveAgentSession, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	sess, ok := t.sessions[agentID]
	//utils.Debug("GetAgentSession: %v", t.sessions)
	return sess, ok
}

// RemoveAgentSession cleans up after disconnect
func (t *EndpointTracker) RemoveAgentSession(agentID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.sessions, agentID)
}

// HasLiveSession checks if an agent has a live session
func (t *EndpointTracker) HasLiveSession(agentID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, ok := t.sessions[agentID]
	return ok
}

func (t *EndpointTracker) GetLiveAgentIDs() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	ids := make([]string, 0, len(t.sessions))
	for id := range t.sessions {
		ids = append(ids, id)
	}
	return ids
}
