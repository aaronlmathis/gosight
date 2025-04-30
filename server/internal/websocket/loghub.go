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

// server/internal/http/websocket/loghub.go
// Description: This file contains the WebSocket hub implementation for the GoSight server.

package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/google/uuid"
)

type LogHub struct {
	clients     map[*Client]bool
	broadcast   chan model.LogPayload
	lock        sync.Mutex
	metaTracker *metastore.MetaTracker
}

func NewLogHub(metaTracker *metastore.MetaTracker) *LogHub {
	return &LogHub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan model.LogPayload, 500), // bursty but medium volume
		metaTracker: metaTracker,
	}
}

func (h *LogHub) Run(ctx context.Context) {
	for {
		select {
		case payload := <-h.broadcast:
			data, _ := json.Marshal(payload)

			h.lock.Lock()

			var deadClients []*Client // NEW: list to track which clients to remove

			for client := range h.clients {
				if h.shouldDeliver(payload, client) {
					if !safeSend(client, data) {
						utils.Warn("Client send failed, scheduling removal: endpoint=%s agent=%s", client.EndpointID, client.AgentID)
						client.Close()
						deadClients = append(deadClients, client)
					}
				}
			}

			// After loop: safely delete dead clients
			for _, client := range deadClients {
				delete(h.clients, client)
			}

			h.lock.Unlock()

		case <-ctx.Done():
			// Context cancelled — shut down cleanly
			h.lock.Lock()
			for client := range h.clients {
				client.Close()
				delete(h.clients, client)
			}
			h.lock.Unlock()

			utils.Info("WebSocketHub: LogHub shutdown complete")
			return
		}
	}
}

func (h *LogHub) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Authenticate the request
	_, err := gosightauth.GetSessionClaims(r)
	if err != nil {
		utils.Warn("Unauthorized WebSocket attempt: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.Warn("WebSocket upgrade failed: %v", err)
		return
	}

	endpointID := r.URL.Query().Get("endpointID")
	meta, _ := h.metaTracker.Get(endpointID)

	client := &Client{
		ID:         uuid.NewString(),
		Conn:       conn,
		EndpointID: endpointID,
		AgentID:    meta.AgentID,
		Send:       make(chan []byte, 100),
	}

	// Heartbeat: read timeout + pong handler
	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Background ping ticker
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			time.Sleep(30 * time.Second)
			if !safeSend(client, []byte(`{"type":"ping"}`)) {
				client.Close()
				h.lock.Lock()
				delete(h.clients, client)
				h.lock.Unlock()
				return
			}
		}
	}()

	// Register client
	h.lock.Lock()
	h.clients[client] = true
	h.lock.Unlock()

	go client.writePump()
}

func (h *LogHub) Broadcast(payload model.LogPayload) {
	select {
	case h.broadcast <- payload:
	default:
		// drop if full
	}
}

func (h *LogHub) shouldDeliver(payload model.LogPayload, client *Client) bool {
	if client.EndpointID == "" {
		return true // no filter
	}

	if payload.EndpointID == client.EndpointID {
		return true // direct match
	}

	if strings.HasPrefix(payload.EndpointID, "ctr-") &&
		payload.Meta != nil &&
		payload.Meta.AgentID == client.AgentID {
		return true // container log belongs to agent
	}

	return false
}
