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

// server/internal/http/router.go
// Router for HTTPServer
package httpserver

import (
	"net/http"
)

func SetupRoutes(mux *http.ServeMux, staticDir, templateDir, env string) {

	// Serve static assets
	fs := http.FileServer(http.Dir(staticDir))
	mux.Handle("/css/", fs)
	mux.Handle("/js/", fs)
	mux.Handle("/images/", fs)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		RenderIndexPage(w, r, templateDir, env)
	})
	mux.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		RenderAgentsPage(w, r, templateDir, env)
	})
	mux.HandleFunc("/containers", func(w http.ResponseWriter, r *http.Request) {
		RenderContainersPage(w, r, templateDir, env)
	})
	mux.HandleFunc("/api/containers", HandleContainersAPI)
	mux.HandleFunc("/api/agents", HandleAgentsAPI)
	// ...
}
