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
				tracker.UpdateAgent(model.Meta{
					Hostname:  "agent-01",
					PrivateIP: "192.168.1.101",
					Tags:      map[string]string{"zone": "DC-1"},
				})
				tracker.UpdateAgent(model.Meta{
					Hostname:  "agent-02",
					PrivateIP: "192.168.1.102",
					Tags:      map[string]string{"zone": "DC-2"},
				})
				tracker.UpdateAgent(model.Meta{
					Hostname:  "agent-03",
					PrivateIP: "192.168.1.103",
					Tags:      map[string]string{"zone": "Narnia"},
				})
				time.Sleep(5 * time.Second)
			}
		}()
	}

	return tracker, nil
}
