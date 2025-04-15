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
// Model for agent
// shared/model/agent.go

package model

import "time"

type AgentStatus struct {
	Hostname   string            `json:"hostname"`
	IP         string            `json:"ip"`
	OS         string            `json:"os"`
	Arch       string            `json:"arch"`
	Version    string            `json:"version"`
	Labels     map[string]string `json:"labels"`
	LastSeen   time.Time         `json:"-"`
	Status     string            `json:"status,omitempty"`
	Since      string            `json:"since,omitempty"`
	Updated    bool              `json:"-"`
	EndpointID string            `json:"endpoint_id,omitempty"`
	AgentID    string            `json:"agent_id,omitempty"`
}
