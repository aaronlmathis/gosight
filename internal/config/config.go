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

type Threshold struct {
	High float64 `yaml:"high" json:"high"`
	Low  float64 `yaml:"low" json:"low"`
}

type ContainerConfig struct {
	Enabled bool   `yaml:"enabled"`
	Runtime string `yaml:"runtime"`
	Socket  string `yaml:"socket"`
}

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`

	Metrics struct {
		IntervalSeconds int `yaml:"interval_seconds"`
	} `yaml:"metrics"`

	Exporters struct {
		Prometheus bool `yaml:"prometheus"`
		Dashboard  bool `yaml:"dashboard"`
	} `yaml:"exporters"`

	Thresholds map[string]Threshold `yaml:"thresholds"`

	Containers ContainerConfig `yaml:"containers"`
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
