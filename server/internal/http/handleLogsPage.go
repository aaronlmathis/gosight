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

// File: server/internal/http/handleLogsPage.go

package httpserver

import (
	"net/http"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func (s *HttpServer) HandleLogsPage(w http.ResponseWriter, r *http.Request) {
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

	bc := map[string]string{"Log Explorer": ""}

	pageData := s.Tmpl.BuildPageData(user, bc, nil, r.URL.Path, "Log Explorer", nil, permissions)
	err = s.Tmpl.RenderTemplate(w, "layout_dashboard", "dashboard_logs", pageData)

	if err != nil {
		http.Error(w, "template error", 500)
	}

}