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
	broadcast chan model.MetricPayload
	lock      sync.Mutex
}
type Client struct {
	Conn       *websocket.Conn
	EndpointID string
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*Client]bool),
		broadcast: make(chan model.MetricPayload, 100),
	}
}

func (h *Hub) Run() {
	for payload := range h.broadcast {
		data, _ := json.Marshal(payload)
		h.lock.Lock()
		for client := range h.clients {
			utils.Debug("Checking client: %p (client.EndpointID=%q, payload.EndpointID=%q)", client, client.EndpointID, payload.EndpointID)

			if client.EndpointID != payload.EndpointID {
				continue // skip non-matching clients
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

func (h *Hub) Broadcast(payload model.MetricPayload) {
	select {
	case h.broadcast <- payload:
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
