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

// server/internal/http/websocket/eventhub.go
// Description: This file contains the WebSocket hub implementation for the GoSight server.

package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type EventsHub struct {
	clients     map[*Client]bool
	broadcast   chan model.EventEntry
	lock        sync.Mutex
	metaTracker *metastore.MetaTracker
}

func NewEventsHub(metaTracker *metastore.MetaTracker) *EventsHub {
	return &EventsHub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan model.EventEntry, 300), // bursty but important
		metaTracker: metaTracker,
	}
}

func (h *EventsHub) Run(ctx context.Context) {
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
			utils.Info("WebSocketHub: EventHub shutdown complete")
			return
		}
	}
}

func (h *EventsHub) ServeWS(w http.ResponseWriter, r *http.Request) {
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
		HostID:     meta.HostID,
		Send:       make(chan []byte, 100),
	}

	// Set up ping/pong heartbeat
	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			time.Sleep(30 * time.Second)
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
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

func (h *EventsHub) Broadcast(payload model.EventEntry) {
	select {
	case h.broadcast <- payload:
	default:
		// drop if full
	}
}

func (h *EventsHub) shouldDeliver(payload model.EventEntry, client *Client) bool {
	if client.EndpointID == "" {
		return true // No filter
	}

	// Primary: Match exact EndpointID
	if payload.EndpointID != "" && payload.EndpointID == client.EndpointID {
		return true
	}
	// Secondary: Match on meta fields
	if payload.Meta != nil {
		if eid, ok := payload.Meta["endpoint_id"]; ok && eid == client.EndpointID {
			return true
		}
		if aid, ok := payload.Meta["agent_id"]; ok && aid == client.AgentID {
			return true
		}
		// New fallback: match if event's meta.host_id == client's endpointID
		if hid, ok := payload.Meta["host_id"]; ok && hid == client.HostID {
			return true
		}
	}

	return false
}
