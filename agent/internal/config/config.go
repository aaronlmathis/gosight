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
// internal/config/config.go
// Package config provides functions to load and manage the configuration for the GoSight agent.
// It handles loading the configuration from a YAML file, applying environment variable overrides,
// and parsing CSV strings into slices. The configuration includes settings for TLS, logging,
// Podman and Docker integration, custom tags, and various collection intervals for metrics,
// logs, and processes. The configuration is structured to allow for easy modification and
// extension as needed.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// LogCollectionConfig defines the configuration for log collection
// It includes settings for the collection interval, sources, services,
// batch size, buffer size, number of workers, and maximum message size.

type LogCollectionConfig struct {
	Interval   time.Duration `yaml:"interval"`
	Sources    []string      `yaml:"sources"`
	Services   []string      `yaml:"services"`
	BatchSize  int           `yaml:"batch_size"`
	BufferSize int           `yaml:"buffer_size"`
	Workers    int           `yaml:"workers"`
	MessageMax int           `yaml:"message_max"`
}

// MetricCollectionConfig defines the configuration for metric collection
// It includes settings for the collection interval, sources, and number of workers.
// The sources can be a list of metrics to collect, such as CPU, memory, etc.
type MetricCollectionConfig struct {
	Interval time.Duration `yaml:"interval"`
	Sources  []string      `yaml:"sources"`
	Workers  int           `yaml:"workers"`
}

// ProcessCollectionConfig defines the configuration for process collection
// It includes settings for the collection interval and number of workers.
// The process collection can be used to monitor running processes and their resource usage.
type ProcessCollectionConfig struct {
	Interval time.Duration `yaml:"interval"`
	Workers  int           `yaml:"workers"`
}

// Config holds the configuration for the GoSight agent.
// It includes settings for TLS, logging, Podman and Docker integration,
// custom tags, and various collection intervals for metrics, logs, and processes.
// The configuration is loaded from a YAML file and can be overridden by environment variables.
// The configuration is structured to allow for easy modification and extension as needed.
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
		ServerURL    string        `yaml:"server_url"`
		Interval     time.Duration `yaml:"interval"`
		HostOverride string        `yaml:"host"`

		MetricCollection  MetricCollectionConfig  `yaml:"metric_collection"`
		LogCollection     LogCollectionConfig     `yaml:"log_collection"`
		ProcessCollection ProcessCollectionConfig `yaml:"process_collection"`

		Environment   string `yaml:"environment"`
		AppLogFile    string `yaml:"app_log_file"`
		ErrorLogFile  string `yaml:"error_log_file"`
		AccessLogFile string `yaml:"access_log_file"`
		DebugLogFile  string `yaml:"debug_log_file"`
		LogLevel      string `yaml:"log_level"`
	}
}

// LoadConfig loads the configuration from a YAML file.
// It returns a Config struct and an error if any occurred during loading.
// The configuration file path is passed as an argument.
// The function reads the file, unmarshals the YAML data into the Config struct,
// and returns the struct. If the file cannot be read or the data cannot be unmarshaled,
// an error is returned.
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

// ApplyEnvOverrides applies environment variable overrides to the configuration.
// It checks for specific environment variables and updates the corresponding fields
// in the Config struct. If an environment variable is set, it overrides the value
// in the configuration. The function prints the overridden values to the console.
// This allows users to customize the configuration without modifying the YAML file.
// The function handles various settings such as server URL, collection intervals,
// log file paths, TLS certificates, and custom tags. It also validates the format
// of the GOSIGHT_INTERVAL environment variable to ensure it is a valid duration.
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

// SplitCSV splits a CSV string into a slice of strings.
// It trims whitespace from each element and ignores empty elements.
// This function is useful for parsing comma-separated values from configuration files.
func SplitCSV(input string) []string {
	var out []string
	for _, s := range strings.Split(input, ",") {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
