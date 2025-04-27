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

// server/internal/http/websocket/hub.go
// Description: This file contains the WebSocket hub implementation for the GoSight server.

package websocket

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/websocket"
)

// Hub manages WebSocket connections and broadcasts.
type Hub struct {
	clients     map[*Client]bool
	broadcast   chan BroadcastEnvelope
	lock        sync.Mutex
	MetaTracker *metastore.MetaTracker
}
type Client struct {
	Conn       *websocket.Conn
	EndpointID string
	AgentID    string
}

type BroadcastEnvelope struct {
	Type string      `json:"type"` // "metrics" or "logs"
	Data interface{} `json:"data"`
}

// NewHub creates a new Hub instance.
// It initializes the clients map and the broadcast channel.
func NewHub(metaTracker *metastore.MetaTracker) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan BroadcastEnvelope, 100),
		MetaTracker: metaTracker,
	}
}

// Run starts the hub's main loop.
// It listens for broadcast messages and sends them to all connected clients.
// The function uses a mutex lock to ensure thread safety when accessing the clients map.
// The broadcast messages are serialized into JSON format before being sent.

func (h *Hub) Run() {

	for envelope := range h.broadcast {

		data, _ := json.Marshal(envelope)

		h.lock.Lock()
		for client := range h.clients {
			// Only filter if the client registered an endpoint
			if client.EndpointID != "" {
				switch envelope.Type {
				case "metrics":
					if payload, ok := envelope.Data.(*model.MetricPayload); ok {
						//utils.Debug("Filtering for client: %s | payload from: %s | host_id: %s",
						//	client.EndpointID, payload.EndpointID, payload.Meta.Tags["host_id"])
						// Exact match (host or container directly watched)
						if payload.EndpointID == client.EndpointID {
							//  direct match
						} else if strings.HasPrefix(payload.EndpointID, "ctr-") &&
							payload.Meta != nil &&
							payload.Meta.Tags != nil &&
							payload.Meta.AgentID == client.AgentID {

						} else {
							continue
						}
					}
				case "logs":
					raw, _ := json.Marshal(envelope.Data)
					var payload model.LogPayload
					if err := json.Unmarshal(raw, &payload); err != nil {
						continue
					}
					if payload.EndpointID != client.EndpointID {
						continue
					}
				}
			}

			err := client.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				client.Conn.Close()
				delete(h.clients, client)
			}
		}
		h.lock.Unlock()
	}
}

// BroadcastMetric sends a metric payload to all connected clients.
// It filters the metric based on the client's endpoint ID and agent ID.
// The metric payload is serialized into JSON format before being sent.
func (h *Hub) BroadcastMetric(payload model.MetricPayload) {
	data, _ := json.Marshal(BroadcastEnvelope{
		Type: "metrics",
		Data: payload,
	})

	h.lock.Lock()
	defer h.lock.Unlock()

	for client := range h.clients {
		if h.shouldDeliver(payload.EndpointID, payload.Meta, client) {
			client.Conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

// BroadcastLog sends a log payload to all connected clients.
// It filters the log based on the client's endpoint ID and agent ID.
// The log payload is serialized into JSON format before being sent.
// The function uses a mutex lock to ensure thread safety when accessing the clients map.
func (h *Hub) BroadcastLog(payload model.LogPayload) {
	data, _ := json.Marshal(BroadcastEnvelope{
		Type: "logs",
		Data: payload,
	})

	h.lock.Lock()
	defer h.lock.Unlock()

	for client := range h.clients {
		if h.shouldDeliver(payload.EndpointID, payload.Meta, client) {
			client.Conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

// BroadcastEvent sends an event to all connected clients.
// It filters the event based on the client's endpoint ID and agent ID.
func (h *Hub) BroadcastEvent(event model.EventEntry) {
	data, _ := json.Marshal(BroadcastEnvelope{
		Type: "event",
		Data: event,
	})

	h.lock.Lock()
	defer h.lock.Unlock()

	for client := range h.clients {

		// 1. Unfiltered client (e.g., overview tab)
		if client.EndpointID == "" {
			_ = client.Conn.WriteMessage(websocket.TextMessage, data)
			continue
		}

		// 2. Direct match: target matches this client
		if event.Target == client.EndpointID {
			_ = client.Conn.WriteMessage(websocket.TextMessage, data)
			continue
		}

		// 3. Indirect match: event is for a container on this host
		if strings.HasPrefix(event.Target, "ctr-") &&
			event.Meta != nil &&
			event.Meta["agent_id"] == client.AgentID {

			_ = client.Conn.WriteMessage(websocket.TextMessage, data)
			continue
		}
	}
}

// BroadcastAlert sends an alert instance to all connected clients.
// It filters the alert based on the client's endpoint ID and agent ID.
// The alert instance is serialized into JSON format before being sent.

func (h *Hub) BroadcastAlert(alert model.AlertInstance) {
	data, _ := json.Marshal(BroadcastEnvelope{
		Type: "alert",
		Data: alert,
	})

	h.lock.Lock()
	defer h.lock.Unlock()

	for client := range h.clients {
		// 1. Unfiltered client (e.g., overview dashboards)
		if client.EndpointID == "" {
			_ = client.Conn.WriteMessage(websocket.TextMessage, data)
			continue
		}

		// 2. Global alert (broadcast to everyone)
		if alert.Scope == "global" {
			_ = client.Conn.WriteMessage(websocket.TextMessage, data)
			continue
		}

		// 3. Direct match: alert is for this client
		if alert.Target == client.EndpointID {
			_ = client.Conn.WriteMessage(websocket.TextMessage, data)
			continue
		}

		// 4. Container match (client is host of container)
		if strings.HasPrefix(alert.Target, "ctr-") &&
			alert.Labels != nil &&
			alert.Labels["agent_id"] == client.AgentID {

			_ = client.Conn.WriteMessage(websocket.TextMessage, data)
			continue
		}
	}
}

// shouldDeliver checks if the payload should be delivered to the client.
// It filters based on the client's endpoint ID and agent ID.
// The function returns true if the payload should be delivered, false otherwise.
func (h *Hub) shouldDeliver(payloadID string, meta *model.Meta, client *Client) bool {
	if client.EndpointID == "" {
		return true // no filtering
	}

	if payloadID == client.EndpointID {
		return true // direct match
	}

	if strings.HasPrefix(payloadID, "ctr-") &&
		meta != nil &&
		meta.AgentID == client.AgentID {
		return true // container belongs to host
	}

	return false
}
func (h *Hub) Broadcast(envelope BroadcastEnvelope) {
	select {
	case h.broadcast <- envelope:
	default:
		// drop message if channel full
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins for dev; tighten this in prod
	},
}

// ServeWS handles WebSocket connections.
// It upgrades the HTTP connection to a WebSocket connection and adds the client to the hub.
// The client is identified by the endpoint ID passed in the URL query parameters.
// The function also retrieves the agent ID from the meta tracker based on the endpoint ID.

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	endpointID := r.URL.Query().Get("endpointID")

	// Pull Meta for that endpoint to link AgentID with container agent_id
	meta, ok := h.MetaTracker.Get(endpointID)
	if !ok {
		utils.Warn("No meta found for endpoint: %s", endpointID)
	}

	client := &Client{
		Conn:       conn,
		EndpointID: endpointID,
		AgentID:    meta.AgentID,
	}

	h.lock.Lock()
	h.clients[client] = true
	h.lock.Unlock()
}
