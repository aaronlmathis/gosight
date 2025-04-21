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

type LogCollectionConfig struct {
	Sources    []string `yaml:"sources"`
	Services   []string `yaml:"services"`
	BatchSize  int      `yaml:"batch_size"`
	MessageMax int      `yaml:"message_max"`
	CursorFile string   `yaml:"cursor_file"`
	LastCursor string   `yaml:"-"` // this field is set dynamically, not from YAML
}

type Config struct {
	TLS struct {
		CAFile   string `yaml:"ca_file"`   // used by agent to trust the server
		CertFile string `yaml:"cert_file"` // optional (for mTLS)
		KeyFile  string `yaml:"key_file"`  // optional (for mTLS)
	}

	Logs struct {
		ErrorLogFile  string `yaml:"error_log_file"`
		AppLogFile    string `yaml:"app_log_file"`
		AccessLogFile string `yaml:"access_log_file"`
		LogLevel      string `yaml:"log_level"`
	}

	Podman struct {
		Socket string `yaml:"socket"`
	}

	Docker struct {
		Socket string `yaml:"socket"`
	}

	Agent struct {
		ServerURL      string              `yaml:"server_url"`
		Interval       time.Duration       `yaml:"interval"`
		HostOverride   string              `yaml:"host"`
		MetricsEnabled []string            `yaml:"metrics_enabled"`
		LogCollection  LogCollectionConfig `yaml:"log_collection"`
		Environment    string              `yaml:"environment"`
		AppLogFile     string              `yaml:"app_log_file"`
		ErrorLogFile   string              `yaml:"error_log_file"`
		AccessLogFile  string              `yaml:"access_log_file"`
		LogLevel       string              `yaml:"log_level"`
		CustomTags     map[string]string   `yaml:"custom_tags"` // static tags to be sent with every metric
	}
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ApplyEnvOverrides(cfg *Config) {
	if val := os.Getenv("GOSIGHT_SERVER_URL"); val != "" {
		cfg.Agent.ServerURL = val
	}
	if val := os.Getenv("GOSIGHT_INTERVAL"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.Agent.Interval = d
		}
	}
	if val := os.Getenv("GOSIGHT_HOST"); val != "" {
		cfg.Agent.HostOverride = val
	}
	if val := os.Getenv("GOSIGHT_METRICS"); val != "" {
		// Comma-separated list
		cfg.Agent.MetricsEnabled = SplitCSV(val)
	}
	if val := os.Getenv("GOSIGHT_ENVIRONMENT"); val != "" {
		cfg.Agent.Environment = val
	}
	if val := os.Getenv("GOSIGHT_ERROR_LOG_FILE"); val != "" {
		cfg.Logs.ErrorLogFile = val
	}
	if val := os.Getenv("GOSIGHT_APP_LOG_FILE"); val != "" {
		cfg.Logs.ErrorLogFile = val
	}
	if val := os.Getenv("GOSIGHT_ACCESS_LOG_FILE"); val != "" {
		cfg.Logs.ErrorLogFile = val
	}
	if val := os.Getenv("GOSIGHT_LOG_LEVEL"); val != "" {
		cfg.Logs.LogLevel = val
	}
	if val := os.Getenv("GOSIGHT_TLS_CERT_FILE"); val != "" {
		cfg.TLS.CertFile = val
	}
	if val := os.Getenv("GOSIGHT_TLS_KEY_FILE"); val != "" {
		cfg.TLS.KeyFile = val
	}
	if val := os.Getenv("GOSIGHT_TLS_CA_FILE"); val != "" {
		cfg.TLS.CAFile = val
	}
	if val := os.Getenv("GOSIGHT_CUSTOM_TAGS"); val != "" {
		// Comma-separated list of key=value pairs
		tags := SplitCSV(val)
		for _, tag := range tags {
			parts := strings.SplitN(tag, "=", 2)
			if len(parts) == 2 {
				cfg.Agent.CustomTags[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
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
