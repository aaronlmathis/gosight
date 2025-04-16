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

func NewHub(metaTracker *metastore.MetaTracker) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan BroadcastEnvelope, 100),
		MetaTracker: metaTracker,
	}
}

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
						utils.Debug("🔎 Filtering for client: %s | payload from: %s | host_id: %s",
							client.EndpointID, payload.EndpointID, payload.Meta.Tags["host_id"])
						// Exact match (host or container directly watched)
						if payload.EndpointID == client.EndpointID {
							// ✅ direct match
						} else if strings.HasPrefix(payload.EndpointID, "container-") &&
							payload.Meta != nil &&
							payload.Meta.Tags != nil &&
							payload.Meta.AgentID == client.AgentID {

						} else {
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

func (h *Hub) shouldDeliver(payloadID string, meta *model.Meta, client *Client) bool {
	if client.EndpointID == "" {
		return true // no filtering
	}

	if payloadID == client.EndpointID {
		return true // direct match
	}

	if strings.HasPrefix(payloadID, "container-") &&
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
