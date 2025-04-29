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
	"os"
	"os/exec"
	"path/filepath"
	"time"

	pb "github.com/aaronlmathis/gosight/shared/proto"
)

// runShellCommand executes a shell command with arguments and returns the result.
func runShellCommand(cmd string, args ...string) (*pb.CommandResponse, error) {
	allowedCommands := []string{"docker", "podman", "systemctl", "ls", "uptime", "reboot", "shutdown"}
	allowed := false
	for _, allowedCmd := range allowedCommands {
		if cmd == allowedCmd {
			allowed = true
			break
		}
	}
	if !allowed {
		return &pb.CommandResponse{Success: false, ErrorMessage: "command not allowed"}, nil
	}

	command := exec.Command(cmd, args...)
	out, err := command.CombinedOutput()

	return &pb.CommandResponse{
		Success: err == nil,
		Output:  string(out),
		ErrorMessage: func() string {
			if err != nil {
				return err.Error()
			}
			return ""
		}(),
	}, nil
}

// runAnsiblePlaybook executes an Ansible playbook from a string and returns the result.
func runAnsiblePlaybook(playbookContent string) (*pb.CommandResponse, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "gosight-playbook-"+time.Now().Format("20060102-150405")+".yml")

	err := os.WriteFile(tmpFile, []byte(playbookContent), 0644)
	if err != nil {
		return &pb.CommandResponse{Success: false, ErrorMessage: "failed to write playbook: " + err.Error()}, nil
	}
	defer os.Remove(tmpFile)

	cmd := exec.Command("ansible-playbook", tmpFile)
	out, err := cmd.CombinedOutput()

	return &pb.CommandResponse{
		Success: err == nil,
		Output:  string(out),
		ErrorMessage: func() string {
			if err != nil {
				return err.Error()
			}
			return ""
		}(),
	}, nil
}
