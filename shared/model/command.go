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

import "time"

type CommandRequest struct {
	AgentID     string   `json:"agent_id"`       // Unique identifier for the agent
	CommandType string   `json:"command_type"`   // e.g., "shell" or "ansible"
	CommandData string   `json:"command_data"`   // The actual shell command or playbook content
	Args        []string `json:"args,omitempty"` // Optional arguments for shell commands
}

type CommandResult struct {
	EndpointID   string `json:"endpoint_id"`
	Output       string `json:"output"`
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_message,omitempty"`
	Timestamp    string `json:"timestamp"` // ISO format
}

type CommandResponse struct {
	Success      bool
	Output       string
	ErrorMessage string
}

type ShellCommandResult struct {
	Output       string        // Full stdout and stderr combined
	ExitCode     int           // Actual exit code (0 = success)
	ErrorMessage string        // Error description if command failed (e.g. exec errors)
	Duration     time.Duration // How long it took to run
	StartedAt    time.Time     // Optional: when the command started
}
type AnsibleCommandResult struct {
	Output       string
	ExitCode     int
	ErrorMessage string
	Duration     time.Duration
	StartedAt    time.Time
	Changed      bool                   // was something changed?
	Stats        map[string]interface{} // from parsed play recap
}
