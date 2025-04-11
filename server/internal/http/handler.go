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

// Basic Handler for http server
// server/internal/http/handler.go
package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/aaronlmathis/gosight/server/internal/contextutil"
)

func FakeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, _ := contextutil.GetUserID(ctx)
	roles, _ := contextutil.GetUserRoles(ctx)
	perms, _ := contextutil.GetUserPermissions(ctx)
	traceID, _ := contextutil.GetTraceID(ctx)

	resp := map[string]interface{}{
		"message":     "✅ You accessed a protected test route!",
		"user_id":     userID,
		"roles":       roles,
		"permissions": perms,
		"trace_id":    traceID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
