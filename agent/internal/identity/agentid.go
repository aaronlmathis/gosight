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

// agent/internal/identity/agentid.go

package agentidentity

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/google/uuid"
)

// LoadOrCreateAgentID returns a stable UUID stored on disk.
// It creates a new one on first run and saves it to disk.
func LoadOrCreateAgentID() (string, error) {
	path := getAgentIDPath()

	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		utils.Debug("Loaded agent ID from disk: %s", string(data))
		return string(data), nil
	}

	id := uuid.NewString()
	utils.Debug("Generated new agent ID: %s", id)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(id), 0600); err != nil {
		return "", err
	}

	return id, nil
}
func getAgentIDPath() string {
	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("APPDATA") // e.g. C:\Users\you\AppData\Roaming
		if base == "" {
			base = "C:\\gosight"
		}
		return filepath.Join(base, "gosight", "agent_id")
	default:
		base := os.Getenv("XDG_STATE_HOME")
		if base == "" {
			base = filepath.Join(os.Getenv("HOME"), ".local", "state")
		}
		return filepath.Join(base, "gosight", "agent_id")
	}
}
