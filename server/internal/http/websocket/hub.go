package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/gorilla/websocket"
)

// Hub manages WebSocket connections and broadcasts.
type Hub struct {
	clients   map[*websocket.Conn]bool
	broadcast chan model.MetricPayload
	lock      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan model.MetricPayload, 100),
	}
}

func (h *Hub) Run() {
	for payload := range h.broadcast {
		data, _ := json.Marshal(payload)
		h.lock.Lock()
		for conn := range h.clients {
			err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				conn.Close()
				delete(h.clients, conn)
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

	h.lock.Lock()
	h.clients[conn] = true
	h.lock.Unlock()
}
