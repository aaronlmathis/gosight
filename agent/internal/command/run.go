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
	"runtime"
	"strings"
	"time"

	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
	"github.com/aaronlmathis/gosight/shared/proto"
)

// runShellCommand executes a shell command with arguments and returns the result.
// It uses a strict allowlist to prevent execution of unapproved commands.
// The command and its arguments are validated for unsafe characters.
// The command is executed in a context, and the output is captured.
func runShellCommand(ctx context.Context, cmd string, args ...string) *proto.CommandResponse {
	var allowed map[string]bool
	osType := runtime.GOOS

	// Strict allowlists
	allowedLinux := map[string]bool{
		"docker": true, "podman": true, "uptime": true, "ls": true, "pwd": true,
		"cat": true, "echo": true, "cp": true, "mv": true, "grep": true,
		"find": true, "chmod": true, "chown": true, "kill": true, "ps": true,
		"systemctl": true, "journalctl": true,
	}
	allowedWindows := map[string]bool{
		"docker": true, "podman": true, "Get-Process": true, "Get-Service": true,
		"Get-ChildItem": true, "Set-Location": true, "Get-Location": true,
	}

	if osType == "windows" {
		allowed = allowedWindows
	} else {
		allowed = allowedLinux
	}

	// Disallow unapproved commands
	if !allowed[cmd] {
		msg := fmt.Sprintf("command not allowed: %s. Allowed: %v", cmd, agentutils.Keys(allowed))
		return &proto.CommandResponse{Success: false, ErrorMessage: msg}
	}

	// Validate all args for dangerous characters
	for _, a := range args {
		if strings.ContainsAny(a, "&|;$><`\\") {
			return &proto.CommandResponse{Success: false, ErrorMessage: "invalid characters in arguments"}
		}
	}

	var execCmd *exec.Cmd

	if osType == "windows" {
		// Build sanitized command line
		full := append([]string{cmd}, args...)
		cmdline := strings.Join(full, " ")

		// Extra check (redundant but safe)
		if strings.ContainsAny(cmdline, "&|;$><`") {
			return &proto.CommandResponse{Success: false, ErrorMessage: "unsafe characters in command line"}
		}

		// Use powershell with -NoProfile -Command
		execCmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", cmdline)
	} else {
		// Directly run the command on Unix
		execCmd = exec.CommandContext(ctx, cmd, args...)
	}

	// Run the command and capture output
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
// It writes the playbook content to a temporary file, executes it, and captures the output.
// The playbook is expected to be in YAML format. The function handles errors related to file writing and command execution.
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
