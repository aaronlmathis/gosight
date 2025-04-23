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

// gosight/agent/internal/bootstrap/web_socket.go

package bootstrap

import (
	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/aaronlmathis/gosight/server/internal/websocket"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// InitWebSocketHub initializes the WebSocket hub for the GoSight agent.
// The WebSocket hub is responsible for managing WebSocket connections and
// broadcasting messages to connected clients.
func InitWebSocketHub(metaStore *metastore.MetaTracker) *websocket.Hub {
	ws := websocket.NewHub(metaStore)
	// Start WebSocket server

	go func() {
		utils.Info("Starting WebSocket hub...")
		ws.Run() // no error returned, but safe to log around
	}()
	return ws
}
