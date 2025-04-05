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
	"net/http"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
)

func StartHTTPServer(cfg *config.Config, tracker *store.AgentTracker, metricStore store.MetricStore, metricIndex *store.MetricIndex) {
	InitHandlers(tracker)

	router := mux.NewRouter()
	apiStore := &APIMetricStore{Store: metricStore}
	SetupRoutes(router, metricIndex, apiStore, cfg.Web.StaticDir, cfg.Web.TemplateDir, cfg.Server.Environment)

	utils.Info("🌐 HTTP server running at %s", cfg.Server.HTTPAddr)
	if err := http.ListenAndServe(cfg.Server.HTTPAddr, router); err != nil {
		utils.Error("HTTP server failed: %v", err)
	}

}
