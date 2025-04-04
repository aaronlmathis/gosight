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

const defaultAgentYAML = `server_url: "localhost:50051"
interval: 2s
host: "dev-machine-01-from config"
metrics_enabled:
  - cpu
  - mem
  - host
  - disk
  - net
  - podman
log_file: "./agent.log"     # Optional — empty means stdout/stderr
log_level: "debug"     # Or "debug", etc.
environment: "dev"   # Or "test", "prod"

tls:
  ca_file: "/certs/ca.crt"
  cert_file: "/certs/client.crt"         # (only needed if doing mTLS)
  key_file: "/certs/client.key"          # (only needed if doing mTLS)
`

func EnsureDefaultConfig(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		return os.WriteFile(path, []byte(defaultAgentYAML), 0644)
	}
	return nil
}
