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

// pack
import (
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/aaronlmathis/gosight/shared/utils"
)

func StartHTTPServer(listenAddr string) {
	utils.Debug("CWD = %s", utils.GetWorkingDir())
	mux := http.NewServeMux()

	// Static files
	webDir := "web"
	fs := http.FileServer(http.Dir(webDir))
	mux.Handle("/css/", fs)
	mux.Handle("/js/", fs)
	mux.Handle("/images/", fs)

	// Root route: render template
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.Debug("🚦 Handler reached: GET /")
		tmplPath := filepath.Join(webDir, "templates", "index.html")
		utils.Debug("🔍 Looking for template at: %s", tmplPath)

		raw, err := os.ReadFile(tmplPath)
		if err != nil {
			utils.Error("❌ Couldn't read file: %v", err)
			http.Error(w, "Failed to read template", http.StatusInternalServerError)
			return
		}
		utils.Debug("📄 index.html contents:\n%s", string(raw))
		utils.Debug("🔍 Attempting to read: %s", tmplPath)

		data, err := os.ReadFile(tmplPath)
		if err != nil {
			utils.Error("❌ Could not read template at %s: %v", tmplPath, err)
			http.Error(w, "File read error", http.StatusInternalServerError)
			return
		}

		utils.Debug("📄 index.html raw contents:\n%s", string(data))

		/*tmpl, err := template.New("index").Parse(string(raw))
		if _, err := os.Stat(tmplPath); err != nil {
			utils.Error("❌ os.Stat failed: %v", err)
			http.Error(w, "Template not found: "+tmplPath, http.StatusInternalServerError)
			return
		}*/
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			utils.Error("Template parsing failed: %v", err)
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, map[string]interface{}{
			"Title":   "GoBright",
			"Env":     "DEV",
			"Message": "This is your system admin dashboard.",
		})
		if err != nil {
			utils.Error("Template execution failed: %v", err)
		}
	})

	utils.Info("🌐 HTTP server running at %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		utils.Error("HTTP server failed: %v", err)
	}

}
