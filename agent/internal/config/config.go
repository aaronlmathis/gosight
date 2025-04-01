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
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type TLSConfig struct {
	CAFile   string `yaml:"ca_file"`   // used by agent to trust the server
	CertFile string `yaml:"cert_file"` // optional (for mTLS)
	KeyFile  string `yaml:"key_file"`  // optional (for mTLS)
}

type AgentConfig struct {
	ServerURL      string        `yaml:"server_url"`
	Interval       time.Duration `yaml:"interval"`
	HostOverride   string        `yaml:"host"`
	MetricsEnabled []string      `yaml:"metrics_enabled"`
	LogFile        string        `yaml:"log_file"`
	LogLevel       string        `yaml:"log_level"`
	TLS            TLSConfig     `yaml:"tls"`
}

func LoadConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AgentConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ApplyEnvOverrides(cfg *AgentConfig) {
	if val := os.Getenv("AGENT_SERVER_URL"); val != "" {
		cfg.ServerURL = val
	}
	if val := os.Getenv("AGENT_INTERVAL"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.Interval = d
		}
	}
	if val := os.Getenv("AGENT_HOST"); val != "" {
		cfg.HostOverride = val
	}
	if val := os.Getenv("AGENT_METRICS"); val != "" {
		// Comma-separated list
		cfg.MetricsEnabled = SplitCSV(val)
	}
	if val := os.Getenv("AGENT_LOG_FILE"); val != "" {
		cfg.LogFile = val
	}
	if val := os.Getenv("AGENT_LOG_LEVEL"); val != "" {
		cfg.LogLevel = val
	}
	if val := os.Getenv("AGENT_TLS_CERT_FILE"); val != "" {
		cfg.TLS.CertFile = val
	}
	if val := os.Getenv("AGENT_TLS_KEY_FILE"); val != "" {
		cfg.TLS.KeyFile = val
	}
	if val := os.Getenv("AGENT_TLS_CA_FILE"); val != "" {
		cfg.TLS.CAFile = val
	}

}

func SplitCSV(input string) []string {
	var out []string
	for _, s := range strings.Split(input, ",") {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}