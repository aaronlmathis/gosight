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

type AgentStatus struct {
	Name     string  `json:"name"`
	Status   string  `json:"status"`   // online, offline, idle
	LastSeen string  `json:"lastSeen"` // "3s ago"
	IP       string  `json:"ip"`
	Zone     string  `json:"zone"`
	CPU      float64 `json:"cpu"` // optional for quick charts
}
