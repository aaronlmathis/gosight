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

// File: server/internal/http/handleMetricExplorer.go
// Description: This file contains the HTTP handlers for the GoSight server's metric explorer API.

package httpserver

import (
	"net/http"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// HandleMetricExplorerPage serves the metric explorer page.
// It checks if the user is authenticated and authorized to view the page.
// If the user is not authenticated, they are redirected to the login page.
// If the user is authenticated, the page is rendered with the user's data and breadcrumbs.

func (s *HttpServer) HandleMetricExplorerPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if forbidden, ok := ctx.Value("forbidden").(bool); ok && forbidden {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, ok := contextutil.GetUserID(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := s.Sys.Stores.Users.GetUserWithPermissions(ctx, userID)
	if err != nil {
		utils.Error("Failed to load user %s: %v", userID, err)
		http.Error(w, "failed to load user", http.StatusInternalServerError)
		return
	}

	bc := make(map[string]string, 0)
	bc["Metric Explorer"] = "/metrics"

	perms := gosightauth.FlattenPermissions(user.Roles)

	pageData := *s.Tmpl.BuildPageData(user, bc, nil, r.URL.Path, "Metric Explorer", nil, perms)

	err = s.Tmpl.RenderTemplate(w, "layout_dashboard", "dashboard_metric_explorer", pageData)

	if err != nil {
		utils.Error("Failed to render metric explorer page: %v", err)
		http.Error(w, "template error", 500)
	}
}
