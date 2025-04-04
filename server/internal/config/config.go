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

type Config struct {
	Server struct {
		GRPCAddr    string `yaml:"grpc_addr"`
		HTTPAddr    string `yaml:"http_addr"`
		Environment string `yaml:"environment"`
		LogFile     string `yaml:"log_file"`
		LogLevel    string `yaml:"log_level"`
	} `yaml:"server"`

	Web struct {
		StaticDir    string `yaml:"static_dir"`
		TemplateDir  string `yaml:"template_dir"`
		DefaultTitle string `yaml:"default_title"`
	} `yaml:"web"`

	TLS struct {
		CertFile     string `yaml:"cert_file"`
		KeyFile      string `yaml:"key_file"`
		ClientCAFile string `yaml:"client_ca_file"`
	} `yaml:"tls"`

	Debug struct {
		EnableReflection bool `yaml:"enable_reflection"`
	} `yaml:"debug"`

	Storage struct {
		Engine        string `yaml:"engine"`
		URL           string `yaml:"url"`
		Workers       int    `yaml:"workers"`
		QueueSize     int    `yaml:"queue_size"`
		BatchSize     int    `yaml:"batch_size"`
		BatchTimeout  int    `yaml:"batch_timeout"`
		BatchRetry    int    `yaml:"batch_retry"`
		BatchInterval int    `yaml:"batch_interval"`
	} `yaml:"storage"`
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
	if val := os.Getenv("SERVER_GRPC_LISTEN"); val != "" {
		cfg.Server.GRPCAddr = val
	}
	if val := os.Getenv("SERVER_HTTP_LISTEN"); val != "" {
		cfg.Server.HTTPAddr = val
	}
	if val := os.Getenv("SERVER_LOG_FILE"); val != "" {
		cfg.Server.LogFile = val
	}
	if val := os.Getenv("SERVER_LOG_LEVEL"); val != "" {
		cfg.Server.LogLevel = val
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
