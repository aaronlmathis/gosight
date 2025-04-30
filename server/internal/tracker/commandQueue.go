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
	"github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type CommandQueue struct {
	Pending []*proto.CommandRequest
}

func (t *EndpointTracker) EnqueueCommand(agentID string, cmd *proto.CommandRequest) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.commandQueue == nil {
		t.commandQueue = make(map[string]*CommandQueue)
	}

	q, exists := t.commandQueue[agentID]
	if !exists {
		q = &CommandQueue{}
		t.commandQueue[agentID] = q
	}

	q.Pending = append(q.Pending, cmd)
	utils.Debug("✅ Enqueued command for %s (queue length: %d)", agentID, len(q.Pending))
	return true
}

func (t *EndpointTracker) DequeueCommand(agentID string) *proto.CommandRequest {
	t.mu.Lock()
	defer t.mu.Unlock()

	q, exists := t.commandQueue[agentID]
	if !exists || len(q.Pending) == 0 {
		return nil
	}

	cmd := q.Pending[0]
	q.Pending = q.Pending[1:]
	return cmd
}
