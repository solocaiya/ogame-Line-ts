package ws

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
	sendBufSize    = 256
)

// Client represents a single WebSocket connection.
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	PlayerID string
	done     bool // set under Hub lock before closing Send; prevents send-on-closed-channel panics
}

// checkOrigin validates the request Origin against allowedOrigins.
// Empty allowedOrigins permits all (development mode).
func (h *Hub) checkOrigin(r *http.Request) bool {
	if len(h.allowedOrigins) == 0 {
		return true // dev mode: allow all
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // same-origin requests (non-browser clients)
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := parsed.Host
	if host == "" {
		host = parsed.Path
	}
	for _, allowed := range h.allowedOrigins {
		if host == allowed || origin == allowed {
			return true
		}
	}
	return false
}

// HandleWS upgrades HTTP to WebSocket and registers the client.
func (h *Hub) HandleWS(c *gin.Context, playerID string) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		Hub:      h,
		Conn:     conn,
		Send:     make(chan []byte, sendBufSize),
		PlayerID: playerID,
	}

	// Unregister any existing connection for this player
	h.mu.Lock()
	if old, ok := h.clients[playerID]; ok {
		old.done = true
		close(old.Send)
		delete(h.clients, playerID)
		old.Conn.Close()
	}
	h.mu.Unlock()

	// Register new client
	h.register <- client

	// Start read/write pumps
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error from %s: %v", c.PlayerID, err)
			}
			break
		}
		// Client messages are currently ignored — server is authoritative
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
