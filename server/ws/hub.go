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
	batchSend      chan BatchMessage // batch messages to specific players
	register       chan *Client
	unregister     chan *Client
	allowedOrigins []string // empty = allow all (dev); set for production
	OnConnect    func(playerID string) // called when a client connects
	OnDisconnect func(playerID string) // called when a client disconnects
}

// Message is the envelope for all WebSocket messages.
type Message struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp int64           `json:"timestamp"`
}

// BatchMessage sends multiple events to a specific player in one WS frame.
type BatchMessage struct {
	PlayerID string
	Messages []Message
}

// NewHub creates a new Hub. allowedOrigins restricts WebSocket origins;
// empty slice allows all (development mode).
func NewHub(allowedOrigins []string) *Hub {
	return &Hub{
		clients:        make(map[string]*Client),
		broadcast:      make(chan []byte, 256),
		batchSend:      make(chan BatchMessage, 256),
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
			if h.OnConnect != nil {
				h.OnConnect(client.PlayerID)
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.PlayerID]; ok {
				delete(h.clients, client.PlayerID)
				client.done = true
				close(client.Send)
			}
			h.mu.Unlock()
			if h.OnDisconnect != nil {
				h.OnDisconnect(client.PlayerID)
			}

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

		case batch := <-h.batchSend:
			h.mu.RLock()
			client, ok := h.clients[batch.PlayerID]
			if ok && !client.done {
				bytes, err := json.Marshal(batch.Messages)
				if err == nil {
					select {
					case client.Send <- bytes:
					default:
						// Client send buffer full, skip
					}
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

// BatchSend sends multiple messages to a specific player in one WS frame.
// This reduces frontend render pressure during high-frequency events.
func (h *Hub) BatchSend(playerID string, messages []Message) {
	if len(messages) == 0 {
		return
	}
	h.batchSend <- BatchMessage{PlayerID: playerID, Messages: messages}
}

// IsConnected returns whether a player has an active WebSocket connection.
func (h *Hub) IsConnected(playerID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[playerID]
	return ok
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
