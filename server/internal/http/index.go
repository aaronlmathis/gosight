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

// Handler for index page
// server/internal/http/index.go

package httpserver

import (
	"net/http"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/http/templates"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func HandleIndex(w http.ResponseWriter, r *http.Request, cfg *config.Config, userStore userstore.UserStore) {
	ctx := r.Context()

	// Check for forbidden access first
	if forbidden, ok := ctx.Value("forbidden").(bool); ok && forbidden {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if user is authenticated
	userID, ok := contextutil.GetUserID(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := userStore.GetUserWithPermissions(ctx, userID)
	if err != nil {
		utils.Error("❌ Failed to load user %s: %v", userID, err)
		http.Error(w, "failed to load user", http.StatusInternalServerError)
		return
	}
	utils.Debug("User details: %s %s % s", user.FirstName, user.LastName, user.Email)

	data := map[string]any{
		"User":        user,
		"Breadcrumbs": "Dashboard / Overview",

		"OverviewMetrics": []map[string]string{
			{
				"Label": "CPU Usage",
				"Value": "42%",
				"Unit":  "",
			},
			{
				"Label": "Memory",
				"Value": "6.2 GB",
				"Unit":  "of 16 GB",
			},
			{
				"Label": "Uptime",
				"Value": "3 days",
				"Unit":  "",
			},
		},
	}

	err = templates.RenderTemplate(w, "dashboard/layout", data)
	if err != nil {
		utils.Error("❌ Template error: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}
