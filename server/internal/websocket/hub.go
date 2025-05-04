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
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/gorilla/websocket"
)

var _ = NewCommandHub

type Client struct {
	Conn       *websocket.Conn
	EndpointID string
	AgentID    string
	HostID     string
	Send       chan []byte
	ID         string
	closeOnce  sync.Once
}

type HubManager struct {
	Metrics   *MetricHub
	Logs      *LogHub
	Alerts    *AlertsHub
	Events    *EventsHub
	Commands  *CommandHub
	Processes *ProcessHub
}

// NewHubManager creates a new HubManager with initialized hubs.
func NewHubManager(metaTracker *metastore.MetaTracker) *HubManager {
	return &HubManager{
		Metrics:   NewMetricHub(metaTracker),
		Logs:      NewLogHub(metaTracker),
		Alerts:    NewAlertsHub(metaTracker),
		Events:    NewEventsHub(metaTracker),
		Commands:  NewCommandHub(metaTracker),
		Processes: NewProcessHub(metaTracker),
	}
}

// StartAll starts all hubs in separate goroutines.
func (h *HubManager) StartAll(ctx context.Context) {
	go h.Metrics.Run(ctx)
	go h.Logs.Run(ctx)
	go h.Alerts.Run(ctx)
	go h.Events.Run(ctx)
	go h.Commands.Run(ctx)
	go h.Processes.Run(ctx)
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
	defer c.Close()

	for msg := range c.Send {
		c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.Send)      // signal writePump to exit
		_ = c.Conn.Close() // forcefully close socket if not already closed
	})
}

func safeSend(c *Client, msg []byte) bool {
	defer func() {
		recover() // recover if send panics because channel is closed
	}()

	select {
	case c.Send <- msg:
		return true
	default:
		return false
	}
}
