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

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		GRPCAddr    string `yaml:"grpc_addr"`
		HTTPAddr    string `yaml:"http_addr"`
		Environment string `yaml:"environment"`
	} `yaml:"server"`

	Logs struct {
		ErrorLogFile  string `yaml:"error_log_file"`
		AppLogFile    string `yaml:"app_log_file"`
		AccessLogFile string `yaml:"access_log_file"`
		LogLevel      string `yaml:"log_level"`
	}
	Web struct {
		StaticDir     string   `yaml:"static_dir"`
		TemplateDir   string   `yaml:"template_dir"`
		DefaultTitle  string   `yaml:"default_title"`
		AuthProviders []string `yaml:"auth_providers"`
	} `yaml:"web"`

	TLS struct {
		HttpsCertFile string `yaml:"https_cert_file"`
		HttpsKeyFile  string `yaml:"https_key_file"`
		CertFile      string `yaml:"cert_file"`
		KeyFile       string `yaml:"key_file"`
		ClientCAFile  string `yaml:"client_ca_file"`
	} `yaml:"tls"`

	Debug struct {
		EnableReflection bool `yaml:"enable_reflection"`
	} `yaml:"debug"`

	// TODO - split up storage engine configs.
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

	LogStore struct {
		Engine        string `yaml:"engine"` // file, victoriametric etc
		Dir           string `yaml:"dir"`
		Workers       int    `yaml:"workers"`
		QueueSize     int    `yaml:"queue_size"`
		BatchSize     int    `yaml:"batch_size"`
		BatchTimeout  int    `yaml:"batch_timeout"`
		BatchRetry    int    `yaml:"batch_retry"`
		BatchInterval int    `yaml:"batch_interval"`
	}

	UserStore struct {
		Type     string `yaml:"type"`      // e.g. "postgres", "memory", "ldap"
		DSN      string `yaml:"dsn"`       // e.g. PostgreSQL connection string
		LDAPBase string `yaml:"ldap_base"` // optional: LDAP-specific config
	} `yaml:"userstore"`

	EventStore struct {
		Engine string `yaml:"engine"` // "memory", "json", or "postgres"
		Path   string `yaml:"path"`   // optional path for JSON file
	} `yaml:"eventstore"`

	RuleStore struct {
		Engine string `yaml:"engine"` // "memory", "json", or "postgres"
		Path   string `yaml:"path"`   // optional path for JSON file
	}

	RouteStore struct {
		Path string `yaml:"path"` // path for YAML file
	}

	Auth struct {
		SSOEnabled bool         `yaml:"sso_enabled"`
		MFASecret  string       `yaml:"mfa_secret_key"`
		JWTSecret  string       `yaml:"jwt_secret"`
		Google     GoogleConfig `yaml:"google"`
	}
}

type GoogleConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
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
	if val := os.Getenv("GOSIGHT_GRPC_LISTEN"); val != "" {
		cfg.Server.GRPCAddr = val
	}
	if val := os.Getenv("GOSIGHT_HTTP_LISTEN"); val != "" {
		cfg.Server.HTTPAddr = val
	}
	if val := os.Getenv("GOSIGHT_ERROR_LOG_FILE"); val != "" {
		cfg.Logs.ErrorLogFile = val
	}
	if val := os.Getenv("GOSIGHT_APP_LOG_FILE"); val != "" {
		cfg.Logs.ErrorLogFile = val
	}
	if val := os.Getenv("GOSIGHT_ACCESS_LOG_FILE"); val != "" {
		cfg.Logs.AccessLogFile = val
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
	if val := os.Getenv("GOSIGHT_TLS_CLIENT_CA_FILE"); val != "" {
		cfg.TLS.ClientCAFile = val
	}
	if val := os.Getenv("GOSIGHT_DEBUG_ENABLE_REFLECTION"); val != "" {
		cfg.Debug.EnableReflection = val == "true"
	}
	if val := os.Getenv("GOSIGHT_USERSTORE_TYPE"); val != "" {
		cfg.UserStore.Type = val
	}
	if val := os.Getenv("GOSIGHT_USERSTORE_DSN"); val != "" {
		cfg.UserStore.DSN = val
	}
	if val := os.Getenv("GOSIGHT_USERSTORE_LDAP_BASE"); val != "" {
		cfg.UserStore.LDAPBase = val
	}
}

func (g *GoogleConfig) ToOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.ClientID,
		ClientSecret: g.ClientSecret,
		RedirectURL:  g.RedirectURI,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}
