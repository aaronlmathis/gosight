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

// server/internal/http/handleAlerts.go

package httpserver

import (
	"net/http"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/http/templates"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func (s *HttpServer) HandleAlertsPage(w http.ResponseWriter, r *http.Request) {

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

	permissions := gosightauth.FlattenPermissions(user.Roles)

	pageData := templates.TemplateData{
		Title:       "Alerts",
		User:        user,
		Permissions: permissions,
		Breadcrumbs: []templates.Breadcrumb{
			{Label: "Alerts"},
		},
	}

	err = templates.RenderTemplate(w, "layout_dashboard", "dashboard_alerts", pageData)

	if err != nil {
		http.Error(w, "template error", 500)
	}

}

func (s *HttpServer) HandleAddAlertRulePage(w http.ResponseWriter, r *http.Request) {
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

	permissions := gosightauth.FlattenPermissions(user.Roles)

	pageData := templates.TemplateData{
		Title:       "Add Alert",
		User:        user,
		Permissions: permissions,
		Breadcrumbs: []templates.Breadcrumb{
			{Label: "Alerts", URL: "/alerts"},
			{Label: "Add Alert"},
		},
	}

	err = templates.RenderTemplate(w, "layout_dashboard", "alerts_add_alert", pageData)

	if err != nil {
		http.Error(w, "template error", 500)
	}
}
