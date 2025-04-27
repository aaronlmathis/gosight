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
	"context"
	"net/http"

	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn       *websocket.Conn
	EndpointID string
	AgentID    string
	Send       chan []byte // NEW
}

type HubManager struct {
	Metrics *MetricHub
	Logs    *LogHub
	Alerts  *AlertsHub
	Events  *EventsHub
}

// NewHubManager creates a new HubManager with initialized hubs.
func NewHubManager(metaTracker *metastore.MetaTracker) *HubManager {
	return &HubManager{
		Metrics: NewMetricHub(metaTracker),
		Logs:    NewLogHub(metaTracker),
		Alerts:  NewAlertsHub(metaTracker),
		Events:  NewEventsHub(metaTracker),
	}
}

// StartAll starts all hubs in separate goroutines.
func (h *HubManager) StartAll(ctx context.Context) {
	go h.Metrics.Run(ctx)
	go h.Logs.Run(ctx)
	go h.Alerts.Run(ctx)
	go h.Events.Run(ctx)
}

// shared WebSocket upgrader used by all hubs
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		/*
			origin := r.Header.Get("Origin")

			if origin == "https://yourfrontend.example.com" {
				return true
			}
			return false
		*/
		return true // allow all origins in dev, tighten later TODO : tighten security
	},
}

func (c *Client) writePump() {
	defer func() {
		c.Conn.Close()
	}()

	for msg := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
}
