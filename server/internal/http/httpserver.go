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
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/http/templates"
	"github.com/aaronlmathis/gosight/server/internal/sys"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type HttpServer struct {
	httpServer *http.Server
	Router     *mux.Router
	Sys        *sys.SystemContext
}

// NewServer creates a new HTTP server instance with the provided system context.
// It initializes the router and sets up the server configuration.
func NewServer(sys *sys.SystemContext) *HttpServer {
	router := mux.NewRouter()
	router.StrictSlash(true)
	s := &HttpServer{
		httpServer: &http.Server{
			Addr:    sys.Cfg.Server.HTTPAddr,
			Handler: router,
		},
		Router: router,
		Sys:    sys,
	}

	return s
}

func (s *HttpServer) Start() error {

	utils.Debug("HttpServer Init Check:\n"+
		"   Config Loaded:           %v\n"+
		"   MetricStore:             %T\n"+
		"   MetricIndex:             %T\n"+
		"   MetaTracker:             %T\n"+
		"   Tracker:            %T\n"+
		"   UserStore:               %T\n"+
		"   DataStore:               %T\n"+
		"   EventStore:              %T\n"+
		"   RuleStore:               %T\n"+
		"   RouteStore:              %T\n"+
		"   AlertManager:            %T\n"+
		"   Router Initialized:      %v\n"+
		"   AuthProviders:           %v\n",
		s.Sys.Cfg != nil,
		s.Sys.Stores.Metrics,
		s.Sys.Tele.Index,
		s.Sys.Tele.Meta,
		s.Sys.Tracker,
		s.Sys.Stores.Users,
		s.Sys.Stores.Data,
		s.Sys.Stores.Events,
		s.Sys.Stores.Rules,
		s.Sys.Stores.Actions,
		s.Sys.Tele.Alerts,
		s.Router != nil,
		getAuthProviderKeys(s.Sys.Auth),
	)
	s.setupRoutes()

	err := templates.InitTemplates(s.Sys.Cfg, s.templateFuncs())
	if err != nil {
		utils.Fatal("template init failed: %v", err)
	}
	utils.Info("HTTPS server running at %s", s.Sys.Cfg.Server.HTTPAddr)
	if err := http.ListenAndServeTLS(s.Sys.Cfg.Server.HTTPAddr, s.Sys.Cfg.TLS.HttpsCertFile, s.Sys.Cfg.TLS.HttpsKeyFile, s.Router); err != nil {
		utils.Error("HTTPS server failed: %v", err)
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
		"marshal":       templates.Marshal,
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
		"seq": func(from, to int) []int {
			s := make([]int, to-from+1)
			for i := range s {
				s[i] = from + i
			}
			return s
		},
	}
}

func (s *HttpServer) Shutdown() error {
	utils.Info("Shutting down HTTP server...")

	if err := s.httpServer.Shutdown(s.Sys.Ctx); err != nil {
		utils.Error("HTTP shutdown error: %v", err)
		return err
	}

	utils.Info("HTTP server shut down cleanly")
	return nil
}
