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

// server/internal/http/websocket/commandhub.go
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
)

type CommandHub struct {
	clients     map[*Client]bool
	broadcast   chan *model.CommandResult // command results
	lock        sync.Mutex
	metaTracker *metastore.MetaTracker
}

func NewCommandHub(metaTracker *metastore.MetaTracker) *CommandHub {
	return &CommandHub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan *model.CommandResult, 100),
		metaTracker: metaTracker,
	}
}

func (h *CommandHub) Run(ctx context.Context) {
	for {
		select {
		case result := <-h.broadcast:
			data, _ := json.Marshal(result)

			h.lock.Lock()
			var dead []*Client
			for client := range h.clients {
				if h.shouldDeliver(result, client) {
					if !safeSend(client, data) {
						utils.Warn("Command client send failed, removing: %s", client.EndpointID)
						client.Close()
						dead = append(dead, client)
					}
				}
			}
			for _, c := range dead {
				delete(h.clients, c)
			}
			h.lock.Unlock()

		case <-ctx.Done():
			h.lock.Lock()
			for c := range h.clients {
				c.Close()
				delete(h.clients, c)
			}
			h.lock.Unlock()
			utils.Info("CommandHub shutdown complete")
			return
		}
	}
}

func (h *CommandHub) ServeWS(w http.ResponseWriter, r *http.Request) {
	_, err := gosightauth.GetSessionClaims(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.Warn("Command WebSocket upgrade failed: %v", err)
		return
	}

	endpointID := r.URL.Query().Get("endpointID")
	meta, _ := h.metaTracker.Get(endpointID)

	client := &Client{
		ID:         uuid.NewString(),
		Conn:       conn,
		EndpointID: endpointID,
		AgentID:    meta.AgentID,
		Send:       make(chan []byte, 50),
	}

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
			if !safeSend(client, []byte(`{"type":"ping"}`)) {
				client.Close()
				h.lock.Lock()
				delete(h.clients, client)
				h.lock.Unlock()
				return
			}
		}
	}()

	h.lock.Lock()
	h.clients[client] = true
	h.lock.Unlock()

	go client.writePump()
}

func (h *CommandHub) Broadcast(result *model.CommandResult) {
	select {
	case h.broadcast <- result:
	default:
		// drop if full
	}
}

func (h *CommandHub) shouldDeliver(result *model.CommandResult, c *Client) bool {
	return result.EndpointID == c.EndpointID
}
