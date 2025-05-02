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

// gosight/server/internal/http/handleLabelValuesAPI.go

package httpserver

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/aaronlmathis/gosight/shared/utils"
)

// HandleLabelValues returns all values for a given label key from
func (s *HttpServer) HandleLabelValues(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing 'key' query parameter", http.StatusBadRequest)
		return
	}

	// Get "contains" if it exists, otherwise use an empty string
	contains := r.URL.Query().Get("contains")

	// Get "limit" query param if exists, otherwise default to a reasonable limit
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default limit
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 100 // ensure limit is a positive integer
		}
	}

	// Get label values with optional filtering by "contains"
	values := s.Sys.Tele.Index.GetLabelValues(key, contains)

	// Sort the values alphabetically (case insensitive)
	sort.SliceStable(values, func(i, j int) bool {
		return strings.ToLower(values[i]) < strings.ToLower(values[j])
	})

	// Apply the limit
	if len(values) > limit {
		values = values[:limit]
	}

	utils.JSON(w, http.StatusOK, values)
}
