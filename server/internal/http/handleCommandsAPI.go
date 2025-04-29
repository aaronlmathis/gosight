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

// server/internal/http/handleCommandsAPI.go

package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/aaronlmathis/gosight/shared/model"
	pb "github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func (s *HttpServer) handleCommandsAPI(w http.ResponseWriter, r *http.Request) {
	var req model.CommandRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	session, _ := s.Sys.Tracker.GetAgentSession(req.AgentID)
	if session == nil {
		http.Error(w, "agent not available", http.StatusServiceUnavailable)
		return
	}

	cmdReq := &pb.CommandRequest{
		EndpointId:  req.AgentID,
		CommandType: req.CommandType,
		Command:     req.CommandData,
		Args:        req.Args,
	}

	s.Sys.Tracker.EnqueueCommand(req.AgentID, cmdReq)

	utils.Info("Enqueued command for agent %s: type=%s command=%s", req.AgentID, req.CommandType, req.CommandData)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "queued",
		"message": "command will be delivered to agent shortly",
	})
}
