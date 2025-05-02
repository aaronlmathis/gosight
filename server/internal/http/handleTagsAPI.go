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

// gosight/agent/internal/http/handleTagsAPI.go

package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
)

// HandleGetTags returns all tags for an endpoint
func (s *HttpServer) HandleGetTags(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpointID := vars["endpointID"]
	utils.Debug("HandleGetTags() for endpoint ID: %s", endpointID)

	tags, err := s.Sys.Stores.Data.GetTags(r.Context(), endpointID)
	if err != nil {
		http.Error(w, "Failed to get tags", http.StatusInternalServerError)
		return
	}

	if tags == nil {
		tags = make(map[string]string) // Return an empty map if no tags exist
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tags)
}

// HandleSetTags replaces all tags for an endpoint
func (s *HttpServer) HandleSetTags(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpointID := vars["endpointID"]

	var tags map[string]string
	if err := json.NewDecoder(r.Body).Decode(&tags); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	err := s.Sys.Stores.Data.SetTags(r.Context(), endpointID, tags)
	if err != nil {
		http.Error(w, "Failed to set tags", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandlePatchTags updates or adds individual tags
func (s *HttpServer) HandlePatchTags(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpointID := vars["endpointID"]

	var updates map[string]string
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	existing, err := s.Sys.Stores.Data.GetTags(r.Context(), endpointID)
	if err != nil {
		http.Error(w, "Failed to fetch existing tags", http.StatusInternalServerError)
		return
	}

	for k, v := range updates {
		existing[k] = v
	}

	err = s.Sys.Stores.Data.SetTags(r.Context(), endpointID, existing)
	if err != nil {
		http.Error(w, "Failed to update tags", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleDeleteTag deletes a specific tag key
func (s *HttpServer) HandleDeleteTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpointID := vars["endpointID"]
	tagKey := vars["key"]

	tagKey = strings.TrimSpace(tagKey)
	if tagKey == "" {
		http.Error(w, "Tag key required", http.StatusBadRequest)
		return
	}

	err := s.Sys.Stores.Data.DeleteTag(r.Context(), endpointID, tagKey)
	if err != nil {
		http.Error(w, "Failed to delete tag", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleTagKeys returns all known tag keys
func (s *HttpServer) HandleTagKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := s.Sys.Stores.Data.ListKeys(r.Context())
	if err != nil {
		http.Error(w, "Failed to list tag keys", http.StatusInternalServerError)
		return
	}
	utils.JSON(w, http.StatusOK, keys)

}

// HandleTagValues returns all values for a given key
func (s *HttpServer) HandleTagValues(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	values, err := s.Sys.Stores.Data.ListValues(r.Context(), key)
	if err != nil {
		http.Error(w, "Failed to list tag values", http.StatusInternalServerError)
		return
	}

	// Optional fuzzy filter
	query := strings.ToLower(r.URL.Query().Get("contains"))
	if query != "" {
		filtered := make([]string, 0, len(values))
		for _, val := range values {
			if strings.Contains(strings.ToLower(val), query) {
				filtered = append(filtered, val)
			}
		}
		values = filtered
	}

	utils.JSON(w, http.StatusOK, values)
}
