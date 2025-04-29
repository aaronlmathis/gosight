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

// shared/model/command.go

package model

type CommandRequest struct {
	AgentID     string   `json:"agent_id"`       // Unique identifier for the agent
	CommandType string   `json:"command_type"`   // e.g., "shell" or "ansible"
	CommandData string   `json:"command_data"`   // The actual shell command or playbook content
	Args        []string `json:"args,omitempty"` // Optional arguments for shell commands
}
