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
	"path/filepath"
	"text/template"

	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func RenderDashboard(w http.ResponseWriter, r *http.Request, templateDir string) {
	ctx := r.Context()
	userID, _ := contextutil.GetUserID(ctx)
	roles, _ := contextutil.GetUserRoles(ctx)
	perms, _ := contextutil.GetUserPermissions(ctx)

	data := map[string]any{
		"UserID":          userID,
		"Roles":           roles,
		"Permissions":     perms,
		"HasAdmin":        HasPermission(perms, "gosight:admin:*"),
		"CanViewUsers":    HasPermission(perms, "gosight:admin:users:view"),
		"CanEditSettings": HasPermission(perms, "gosight:admin:settings:edit"),
	}

	tmpl := template.Must(template.ParseFiles(filepath.Join(templateDir, "dashboard.html")))
	tmpl.Execute(w, data)
}

func RenderIndexPage(w http.ResponseWriter, r *http.Request, templateDir, env string) {
	tmplPath := filepath.Join(templateDir, "index.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		utils.Error("Template parse error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title": "GoSight",
		"Env":   env,
	}

	tmpl.Execute(w, data)
}

func HandleIndex(w http.ResponseWriter, r *http.Request, templateDir, env string) {
	layout := filepath.Join(templateDir, "layout.html")
	page := filepath.Join(templateDir, "index.html")
	tmpl, err := template.ParseFiles(layout, page)

	if err != nil {
		http.Error(w, "Template parsing error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout.html", nil)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
	}
}

func staticWithProperMIMEs(fs http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch filepath.Ext(r.URL.Path) {
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		}
		fs.ServeHTTP(w, r)
	})
}
