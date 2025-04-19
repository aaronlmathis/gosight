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

// server/internal/http/httpserver.go
// Basic http server for admin/dash

package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/http/templates"
	"github.com/aaronlmathis/gosight/server/internal/http/websocket"
	"github.com/aaronlmathis/gosight/server/internal/store/agenttracker"
	"github.com/aaronlmathis/gosight/server/internal/store/datastore"
	"github.com/aaronlmathis/gosight/server/internal/store/logstore"
	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/aaronlmathis/gosight/server/internal/store/metricindex"
	"github.com/aaronlmathis/gosight/server/internal/store/metricstore"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type HttpServer struct {
	AgentTracker   *agenttracker.AgentTracker
	APIMetricStore *APIMetricStore
	AuthProviders  map[string]gosightauth.AuthProvider
	Config         *config.Config
	MetricIndex    *metricindex.MetricIndex
	MetricStore    metricstore.MetricStore
	LogStore       logstore.LogStore
	MetaTracker    *metastore.MetaTracker
	Router         *mux.Router
	UserStore      userstore.UserStore
	DataStore      datastore.DataStore
	WebSocket      *websocket.Hub
	httpServer     *http.Server
}

func NewServer(
	ctx context.Context,
	agentTracker *agenttracker.AgentTracker,
	authProviders map[string]gosightauth.AuthProvider,
	cfg *config.Config,
	metaTracker *metastore.MetaTracker,
	metricIndex *metricindex.MetricIndex,
	metricStore metricstore.MetricStore,
	logStore logstore.LogStore,
	userStore userstore.UserStore,
	dataStore datastore.DataStore,
	webSocket *websocket.Hub,
) *HttpServer {
	router := mux.NewRouter()
	router.StrictSlash(true)
	s := &HttpServer{
		AgentTracker:   agentTracker,
		APIMetricStore: &APIMetricStore{Store: metricStore},
		AuthProviders:  authProviders,
		Config:         cfg,
		MetaTracker:    metaTracker,
		MetricIndex:    metricIndex,
		MetricStore:    metricStore,
		Router:         router,
		UserStore:      userStore,
		DataStore:      dataStore,
		WebSocket:      webSocket,
		LogStore:       logStore,
		httpServer: &http.Server{
			Addr:    cfg.Server.HTTPAddr,
			Handler: router,
		},
	}

	return s
}

func (s *HttpServer) Start() error {

	utils.Debug("HttpServer Init Check:\n"+
		"   Config Loaded:           %v\n"+
		"   MetricStore:             %T\n"+
		"   MetricIndex:             %T\n"+
		"   MetaTracker:             %T\n"+
		"   AgentTracker:            %T\n"+
		"   UserStore:               %T\n"+
		"   Router Initialized:      %v\n"+
		"   AuthProviders:           %v\n",
		s.Config != nil,
		s.MetricStore,
		s.MetricIndex,
		s.MetaTracker,
		s.AgentTracker,
		s.UserStore,
		s.Router != nil,
		getAuthProviderKeys(s.AuthProviders),
	)
	s.setupRoutes()

	err := templates.InitTemplates(s.Config, s.templateFuncs())
	if err != nil {
		utils.Fatal("template init failed: %v", err)
	}
	utils.Info("HTTP server running at %s", s.Config.Server.HTTPAddr)
	if err := http.ListenAndServe(s.Config.Server.HTTPAddr, s.Router); err != nil {
		utils.Error("HTTP server failed: %v", err)
		return err
	}
	return nil
}
func getAuthProviderKeys(m map[string]gosightauth.AuthProvider) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
func (s *HttpServer) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"hasPermission": func(_, _ interface{}) bool { return true }, // stub or hook
		"safeHTML":      func(s string) template.HTML { return template.HTML(s) },
		"title":         cases.Title(language.English).String,
		"toJson": func(v interface{}) template.JS {
			b, _ := json.Marshal(v)
			return template.JS(b)
		},
		"since": func(ts string) string {
			t, err := time.Parse(time.RFC3339, ts)
			if err != nil {
				return "unknown"
			}
			d := time.Since(t)
			if d < time.Minute {
				return fmt.Sprintf("%ds ago", int(d.Seconds()))
			}
			if d < time.Hour {
				return fmt.Sprintf("%dm ago", int(d.Minutes()))
			}
			return fmt.Sprintf("%dh ago", int(d.Hours()))
		},
		"uptime": templates.FormatUptime,
		"trim":   strings.TrimSpace,
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}
}

func (s *HttpServer) Shutdown(ctx context.Context) error {
	utils.Info("Shutting down HTTP server...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		utils.Error("HTTP shutdown error: %v", err)
		return err
	}

	utils.Info("HTTP server shut down cleanly")
	return nil
}
