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
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type LogCollectionConfig struct {
	Sources    []string `yaml:"sources"`
	Services   []string `yaml:"services"`
	BatchSize  int      `yaml:"batch_size"`
	BufferSize int      `yaml:"buffer_size"`
	Workers    int      `yaml:"workers"`
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
		Socket  string `yaml:"socket"`
		Enabled bool   `yaml:"enabled"`
	}

	Docker struct {
		Socket  string `yaml:"socket"`
		Enabled bool   `yaml:"enabled"`
	}

	CustomTags map[string]string `yaml:"custom_tags"` // static tags to be sent with every metric

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
		fmt.Printf("Env override: GOSIGHT_SERVER_URL = %s\n", val)
	}
	if val := os.Getenv("GOSIGHT_INTERVAL"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.Agent.Interval = d
			fmt.Printf("Env override: GOSIGHT_INTERVAL = %s\n", val)
		} else {
			fmt.Printf("Invalid GOSIGHT_INTERVAL format: %s\n", val)
		}
	}
	if val := os.Getenv("GOSIGHT_HOST"); val != "" {
		cfg.Agent.HostOverride = val
		fmt.Printf("Env override: GOSIGHT_HOST = %s\n", val)
	}
	if val := os.Getenv("GOSIGHT_METRICS"); val != "" {
		cfg.Agent.MetricsEnabled = SplitCSV(val)
		fmt.Printf("Env override: GOSIGHT_METRICS = %s\n", val)
	}
	if val := os.Getenv("GOSIGHT_ENVIRONMENT"); val != "" {
		cfg.Agent.Environment = val
		fmt.Printf("Env override: GOSIGHT_ENVIRONMENT = %s\n", val)
	}

	// Log paths
	if val := os.Getenv("GOSIGHT_ERROR_LOG_FILE"); val != "" {
		cfg.Logs.ErrorLogFile = val
		fmt.Printf("Env override: GOSIGHT_ERROR_LOG_FILE = %s\n", val)
	}
	if val := os.Getenv("GOSIGHT_APP_LOG_FILE"); val != "" {
		cfg.Logs.AppLogFile = val
		fmt.Printf("Env override: GOSIGHT_APP_LOG_FILE = %s\n", val)
	}
	if val := os.Getenv("GOSIGHT_ACCESS_LOG_FILE"); val != "" {
		cfg.Logs.AccessLogFile = val
		fmt.Printf("Env override: GOSIGHT_ACCESS_LOG_FILE = %s\n", val)
	}
	if val := os.Getenv("GOSIGHT_LOG_LEVEL"); val != "" {
		cfg.Logs.LogLevel = val
		fmt.Printf("Env override: GOSIGHT_LOG_LEVEL = %s\n", val)
	}

	// TLS certs
	if val := os.Getenv("GOSIGHT_TLS_CERT_FILE"); val != "" {
		cfg.TLS.CertFile = val
		fmt.Printf("Env override: GOSIGHT_TLS_CERT_FILE = %s\n", val)
	}
	if val := os.Getenv("GOSIGHT_TLS_KEY_FILE"); val != "" {
		cfg.TLS.KeyFile = val
		fmt.Printf("Env override: GOSIGHT_TLS_KEY_FILE = %s\n", val)
	}
	if val := os.Getenv("GOSIGHT_TLS_CA_FILE"); val != "" {
		cfg.TLS.CAFile = val
		fmt.Printf("Env override: GOSIGHT_TLS_CA_FILE = %s\n", val)
	}

	// Podman socket override
	if val := os.Getenv("GOSIGHT_PODMAN_SOCKET"); val != "" {
		cfg.Podman.Socket = val
		fmt.Printf("Env override: GOSIGHT_PODMAN_SOCKET = %s\n", val)
	}
	// Docker socket override
	if val := os.Getenv("GOSIGHT_DOCKER_SOCKET"); val != "" {
		cfg.Docker.Socket = val
		fmt.Printf("Env override: GOSIGHT_DOCKER_SOCKET = %s\n", val)
	}

	// Custom tags
	if val := os.Getenv("GOSIGHT_CUSTOM_TAGS"); val != "" {
		fmt.Printf("Loading custom tags from GOSIGHT_CUSTOM_TAGS env: %s\n", val)
		if cfg.CustomTags == nil {
			cfg.CustomTags = make(map[string]string)
		}

		tags := strings.Split(val, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			parts := strings.SplitN(tag, "=", 2)
			if len(parts) != 2 {
				fmt.Printf("Invalid custom tag format (skipped): %s\n", tag)
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key == "" || value == "" {
				fmt.Printf("Empty key or value in custom tag (skipped): %s\n", tag)
				continue
			}
			cfg.CustomTags[key] = value
			fmt.Printf("Custom tag loaded: %s=%s\n", key, value)
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
