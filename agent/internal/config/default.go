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

package config

import (
	"os"
	"path/filepath"
)

const defaultAgentYAML = `agent:
  server_url: "localhost:50051"    # domain/ip:port
  interval: 2s              # Metric collection / send interval
  host: "dev-machine-01"    # Hostname of agent machine
  metrics_enabled:          # Enabled collectors (found in agent/internal/collector and loaded from agent/internal/collector/registry.go)
    - cpu
    - mem
    - host
    - disk
    - net
    - podman
    - docker
  log_collection:
    sources:
      - journald
    batch_size:  50     # Number of log entries to send in a payload
    message_max: 1000   # Max size of messages before truncating (like in journald)
  environment: "dev" # (dev/prod)

# Log Config
logs:
  app_log_file: "./agent.log"      # Relative to path of execution
  error_log_file: "error.log"      # Relative to path of execution
  log_level: "debug"               # Or "info", etc.

# TLS Config
tls:
  ca_file: "../certs/ca.crt"
  cert_file: "../certs/client.crt"         # (only needed if doing mTLS)
  key_file: "../certs/client.key"          # (only needed if doing mTLS)

# Podman collector config
podman:
  enabled: false
  socket: "/run/user/1000/podman/podman.sock"

docker:
  enabled: true
  socket: "/var/run/docker.sock"

`

// EnsureDefaultConfig checks if the default config file exists at the specified path.
// If it does not exist, it creates the directory structure and writes the default config to the file.

func EnsureDefaultConfig(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		return os.WriteFile(path, []byte(defaultAgentYAML), 0644)
	}
	return nil
}
