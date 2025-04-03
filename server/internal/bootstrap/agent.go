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

// server/internal/bootstrap/agent.go
// Init agent tracking from in-memory

package bootstrap

import (
	"math/rand"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/shared/model"
)

func InitAgentTracker(env string) (*store.AgentTracker, error) {
	tracker := store.NewAgentTracker()

	if env == "dev" {
		// Simulate agent data in dev mode
		go func() {
			for {
				tracker.UpdateAgent("agent-01", model.AgentStatus{
					Name: "agent-01", IP: "192.168.1.101", Zone: "DC-1", CPU: float64(rand.Intn(60)),
				})
				tracker.UpdateAgent("agent-02", model.AgentStatus{
					Name: "agent-02", IP: "192.168.1.102", Zone: "DC-2", CPU: float64(rand.Intn(50)),
				})
				tracker.UpdateAgent("agent-03", model.AgentStatus{
					Name: "agent-03", IP: "192.168.1.103", Zone: "Edge", CPU: float64(rand.Intn(40)),
				})
				time.Sleep(5 * time.Second)
			}
		}()
	}

	return tracker, nil
}
