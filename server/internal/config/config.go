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

// server/internal/config/config.go

// Package config provides configuration loading and management for the GoSight server.
// It supports loading configuration from a YAML file and allows for environment variable overrides.
// The configuration includes settings for server address, storage engine, database path,
// logging, TLS settings, and debug options.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type TLSConfig struct {
	CertFile     string `yaml:"cert_file"`
	KeyFile      string `yaml:"key_file"`
	ClientCAFile string `yaml:"client_ca_file"` // Optional (for mTLS)
}

type DebugConfig struct {
	EnableReflection bool `yaml:"enable_reflection"`
}

type StorageConfig struct {
	Engine        string `yaml:"engine"`
	URL           string `yaml:"url"`
	Workers       int    `yaml:"workers"`
	QueueSize     int    `yaml:"queue_size"`
	BatchSize     int    `yaml:"batch_size"`
	BatchTimeout  int    `yaml:"batch_timeout"`
	BatchRetry    int    `yaml:"batch_retry"`
	BatchInterval int    `yaml:"batch_interval"`
}

type ServerConfig struct {
	ListenAddr string        `yaml:"listen"`
	LogFile    string        `yaml:"log_file"`
	LogLevel   string        `yaml:"log_level"`
	TLS        TLSConfig     `yaml:"tls"`
	Debug      DebugConfig   `yaml:"debug"`
	Storage    StorageConfig `yaml:"storage"`
}

func LoadConfig(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ServerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func ApplyEnvOverrides(cfg *ServerConfig) {
	if val := os.Getenv("SERVER_LISTEN"); val != "" {
		cfg.ListenAddr = val
	}
	if val := os.Getenv("SERVER_LOG_FILE"); val != "" {
		cfg.LogFile = val
	}
	if val := os.Getenv("SERVER_LOG_LEVEL"); val != "" {
		cfg.LogLevel = val
	}
	if val := os.Getenv("SERVER_TLS_CERT_FILE"); val != "" {
		cfg.TLS.CertFile = val
	}
	if val := os.Getenv("SERVER_TLS_KEY_FILE"); val != "" {
		cfg.TLS.KeyFile = val
	}
	if val := os.Getenv("SERVER_TLS_CLIENT_CA_FILE"); val != "" {
		cfg.TLS.ClientCAFile = val
	}
	if val := os.Getenv("SERVER_DEBUG_ENABLE_REFLECTION"); val != "" {
		cfg.Debug.EnableReflection = val == "true"
	}
}
