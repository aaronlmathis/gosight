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

// agent/internal/command/run.go

package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
	"github.com/aaronlmathis/gosight/shared/proto"
)

// runShellCommand executes a shell command with arguments and returns the result.
func runShellCommand(ctx context.Context, cmd string, args ...string) *proto.CommandResponse {
	allowed := map[string]bool{
		"docker": true, "podman": true, "systemctl": true,
		"ls": true, "uptime": true, "reboot": true, "shutdown": true,
	}
	if !allowed[cmd] {
		msg := fmt.Sprintf("command not allowed: %s. Allowed: %v", cmd, agentutils.Keys(allowed))
		return &proto.CommandResponse{Success: false, ErrorMessage: msg}
	}

	execCmd := exec.CommandContext(ctx, cmd, args...)
	output, err := execCmd.CombinedOutput()

	success := err == nil
	if exitErr, ok := err.(*exec.ExitError); ok {
		success = exitErr.ExitCode() == 0
	}

	return &proto.CommandResponse{
		Success:      success,
		Output:       string(output),
		ErrorMessage: agentutils.ErrMsg(err),
	}
}

// runAnsiblePlaybook executes an Ansible playbook from a string and returns the result.
func runAnsiblePlaybook(ctx context.Context, playbookContent string) *proto.CommandResponse {
	tmpFile := filepath.Join(os.TempDir(), "gosight-playbook-"+time.Now().Format("20060102-150405")+".yml")

	err := os.WriteFile(tmpFile, []byte(playbookContent), 0644)
	if err != nil {
		return &proto.CommandResponse{
			Success:      false,
			ErrorMessage: "failed to write playbook: " + err.Error(),
		}
	}
	defer os.Remove(tmpFile)

	cmd := exec.CommandContext(ctx, "ansible-playbook", tmpFile)
	out, err := cmd.CombinedOutput()

	return &proto.CommandResponse{
		Success:      err == nil,
		Output:       string(out),
		ErrorMessage: agentutils.ErrMsg(err),
	}
}
