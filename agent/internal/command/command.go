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

// agent/internal/command/command.go

package command

import (
	"github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// HandleCommand processes incoming command requests based on their type.
func HandleCommand(cmd *proto.CommandRequest) {
	switch cmd.CommandType {
	case "shell":
		response, err := runShellCommand(cmd.Command, cmd.Args...)
		if err != nil {
			utils.Warn("Shell command execution error: %v", err)
		}
		utils.Info("Shell Command Output: %s", response.Output)
		if response.ErrorMessage != "" {
			utils.Warn("Shell Command Error: %s", response.ErrorMessage)
		}

	case "ansible":
		response, err := runAnsiblePlaybook(cmd.Command)
		if err != nil {
			utils.Warn("Ansible playbook execution error: %v", err)
		}
		utils.Info("Ansible Output: %s", response.Output)
		if response.ErrorMessage != "" {
			utils.Warn("Ansible Error: %s", response.ErrorMessage)
		}

	default:
		utils.Warn("Unknown command type: %s", cmd.CommandType)
	}
}
