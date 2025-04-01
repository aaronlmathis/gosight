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
along with LeetScraper. If not, see https://www.gnu.org/licenses/.
*/

package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	ListenAddr    string `yaml:"listen"`
	StorageEngine string `yaml:"storage"`
	DatabasePath  string `yaml:"database_path"`
	LogFile       string `yaml:"log_file"`
	LogLevel      string `yaml:"log_level"`
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
	if val := os.Getenv("SERVER_STORAGE"); val != "" {
		cfg.StorageEngine = val
	}
	if val := os.Getenv("SERVER_DATABASE_PATH"); val != "" {
		cfg.DatabasePath = val
	}
	if val := os.Getenv("SERVER_LOG_FILE"); val != "" {
		cfg.LogFile = val
	}
}
