package ws

import (
	"encoding/json"
	"sync"
	"time"
)

// Hub manages all WebSocket client connections and message broadcasting.
type Hub struct {
	mu             sync.RWMutex
	clients        map[string]*Client // playerID -> client
	broadcast      chan []byte
	register       chan *Client
	unregister     chan *Client
	allowedOrigins []string // empty = allow all (dev); set for production
}

// Message is the envelope for all WebSocket messages.
type Message struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp int64           `json:"timestamp"`
}

// NewHub creates a new Hub. allowedOrigins restricts WebSocket origins;
// empty slice allows all (development mode).
func NewHub(allowedOrigins []string) *Hub {
	return &Hub{
		clients:        make(map[string]*Client),
		broadcast:      make(chan []byte, 256),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		allowedOrigins: allowedOrigins,
	}
}

// Run starts the hub's main loop. Should be called in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.PlayerID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.PlayerID]; ok {
				delete(h.clients, client.PlayerID)
				client.done = true
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				if client.done {
					continue
				}
				select {
				case client.Send <- message:
				default:
					// Client send buffer full, skip
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(msgType string, data interface{}) {
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}

	msg := Message{
		Type:      msgType,
		Data:      raw,
		Timestamp: time.Now().UnixMilli(),
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.broadcast <- bytes
}

// SendTo sends a message to a specific player.
func (h *Hub) SendTo(playerID string, msgType string, data interface{}) {
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}

	msg := Message{
		Type:      msgType,
		Data:      raw,
		Timestamp: time.Now().UnixMilli(),
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	client, ok := h.clients[playerID]
	if !ok || client.done {
		h.mu.RUnlock()
		return
	}
	select {
	case client.Send <- bytes:
	default:
		// Client send buffer full, skip
	}
	h.mu.RUnlock()
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
