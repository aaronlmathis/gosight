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

// server/internal/http/server.go
// Basic http server for admin/dash

package httpserver

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/http/templates"
	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var AuthProviders map[string]gosightauth.AuthProvider

func StartHTTPServer(cfg *config.Config, tracker *store.AgentTracker, metricStore store.MetricStore, metricIndex *store.MetricIndex, userStore userstore.UserStore, metaTracker *metastore.MetaTracker) {

	// Decode and store MFASecret and JWTSecret
	err := gosightauth.InitJWTSecret(cfg.Auth.JWTSecret)
	if err != nil {
		utils.Error("failed to decode JWT secret: %v", err)
		return
	}

	err = gosightauth.InitMFAKey(cfg.Auth.MFASecret)
	if err != nil {
		utils.Error("failed to decode MFA secret: %v", err)
		return
	}

	InitHandlers(tracker)

	router := mux.NewRouter()
	apiStore := &APIMetricStore{Store: metricStore}

	AuthProviders, err := BuildAuthProviders(cfg, userStore)
	if err != nil {
		log.Fatalf("failed to build auth providers: %v", err)
	}

	SetupRoutes(router, metricIndex, apiStore, userStore, metaTracker, AuthProviders, cfg)

	// Static file server
	staticDir := http.Dir(cfg.Web.StaticDir)
	fs := http.FileServer(staticDir)

	router.PathPrefix("/js/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set MIME type for all .js files
		switch filepath.Ext(r.URL.Path) {
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		}

		// Full path to the file on disk
		fullPath := filepath.Join(cfg.Web.StaticDir, r.URL.Path)

		// Serve it
		http.ServeFile(w, r, fullPath)
	})

	router.PathPrefix("/css/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set MIME type for all .js files
		switch filepath.Ext(r.URL.Path) {
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		}

		// Full path to the file on disk
		fullPath := filepath.Join(cfg.Web.StaticDir, r.URL.Path)

		// Serve it
		http.ServeFile(w, r, fullPath)
	})

	funcMap := template.FuncMap{
		"hasPermission": func(_, _ interface{}) bool { return true },
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
	}

	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", fs))

	// Initialize templates with the function map
	templates.InitTemplates(cfg, funcMap)

	utils.Info("🌐 HTTP server running at %s", cfg.Server.HTTPAddr)
	if err := http.ListenAndServe(cfg.Server.HTTPAddr, router); err != nil {
		utils.Error("HTTP server failed: %v", err)
	}
}
