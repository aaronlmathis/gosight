package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/websocket"
)

// Hub manages WebSocket connections and broadcasts.
type Hub struct {
	clients   map[*Client]bool
	broadcast chan BroadcastEnvelope
	lock      sync.Mutex
}
type Client struct {
	Conn       *websocket.Conn
	EndpointID string
}

type BroadcastEnvelope struct {
	Type string      `json:"type"` // "metrics" or "logs"
	Data interface{} `json:"data"`
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*Client]bool),
		broadcast: make(chan BroadcastEnvelope, 100),
	}
}

func (h *Hub) Run() {
	for envelope := range h.broadcast {
		data, _ := json.Marshal(envelope)

		h.lock.Lock()
		for client := range h.clients {
			// If filtering by endpoint
			if client.EndpointID != "" {
				switch envelope.Type {
				case "metrics":
					if payload, ok := envelope.Data.(*model.MetricPayload); ok {
						if payload.EndpointID != client.EndpointID {
							continue
						}
					}
				case "logs":
					if payload, ok := envelope.Data.(*model.LogPayload); ok {
						if payload.EndpointID != client.EndpointID {
							continue
						}
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

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	endpointID := r.URL.Query().Get("endpointID")
	client := &Client{Conn: conn, EndpointID: endpointID}
	utils.Debug("New WebSocket client connected: %p (endpointID=%q)", client, endpointID)
	h.lock.Lock()
	h.clients[client] = true
	h.lock.Unlock()
}

/*
if client.EndpointID != "" {
	switch envelope.Type {
	case "metrics":
		if payload, ok := envelope.Data.(*model.MetricPayload); ok {
			if payload.EndpointID == client.EndpointID {
				// exact match ✅
			} else if payload.Meta != nil && payload.Meta.Tags != nil {
				// match container payload if its host_id matches client's endpoint
				if payload.Meta.Tags["host_id"] == client.EndpointID {
					// linked container ✅
				} else {
					continue
				}
			} else {
				continue
			}
		}
	}
*/
