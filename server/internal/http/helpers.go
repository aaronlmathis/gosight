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

// server/internal/http/helpers.go
// Description: This file contains helper functions for the GoSight HTTPS server

package httpserver

import (
	"net/http"
	"strings"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/usermodel"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// BuildAuthEventMeta constructs a metadata map for authentication events.
// It includes user information, provider, IP address, trace ID, user agent, and roles.
// This metadata can be used for logging or auditing purposes.

func (s *HttpServer) BuildAuthEventMeta(user *usermodel.User, r *http.Request) map[string]string {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = "unknown"
	}
	ctx := r.Context()
	traceID, _ := contextutil.GetTraceID(ctx)

	userID := ""
	userEmail := ""
	roleNames := ""

	if user != nil {
		userID = user.ID
		userEmail = user.Email
		roleNames = strings.Join(gosightauth.ExtractRoleNames(user.Roles), ",")
	}

	return map[string]string{
		"user_id":    userID,
		"email":      userEmail,
		"provider":   provider,
		"ip":         utils.GetClientIP(r),
		"trace_id":   traceID,
		"user_agent": r.UserAgent(),
		"roles":      roleNames,
	}
}
