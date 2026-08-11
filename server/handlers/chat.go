package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"ogame-server/models"
	"ogame-server/ws"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChatHandler handles chat messaging and DND settings.
type ChatHandler struct {
	db    *sql.DB
	wsHub *ws.Hub

	// Rate limiting: playerID -> last send time
	mu         sync.Mutex
	lastSendAt map[string]time.Time
}

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(db *sql.DB, wsHub *ws.Hub) *ChatHandler {
	return &ChatHandler{db: db, wsHub: wsHub, lastSendAt: make(map[string]time.Time)}
}

const chatRateLimit = 500 * time.Millisecond // min interval between messages per player

// Send sends a chat message to world or alliance channel.
func (h *ChatHandler) Send(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		Channel string `json:"channel" binding:"required"`
		Content string `json:"content" binding:"required,min=1,max=500"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" || len(req.Content) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content must be 1-500 characters"})
		return
	}

	if req.Channel != "world" && req.Channel != "alliance" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel"})
		return
	}

	// Rate limit: max 1 message per 500ms per player
	h.mu.Lock()
	if last, ok := h.lastSendAt[playerID]; ok && time.Since(last) < chatRateLimit {
		h.mu.Unlock()
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "please slow down"})
		return
	}
	h.lastSendAt[playerID] = time.Now()
	h.mu.Unlock()

	// Get sender info
	var senderName string
	h.db.QueryRow("SELECT COALESCE(NULLIF(display_name,''), username) FROM users WHERE id = ?", playerID).Scan(&senderName)

	// For alliance channel, verify membership and get tag
	var senderTag string
	if req.Channel == "alliance" {
		var allianceID, allianceTag string
		err := h.db.QueryRow(
			`SELECT am.alliance_id, a.tag FROM alliance_members am
			 JOIN alliances a ON am.alliance_id = a.id
			 WHERE am.player_id = ?`, playerID,
		).Scan(&allianceID, &allianceTag)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "not in an alliance"})
			return
		}
		senderTag = allianceTag
	}

	msg := models.ChatMessage{
		ID:         uuid.New().String(),
		Channel:    req.Channel,
		SenderID:   playerID,
		SenderName: senderName,
		SenderTag:  senderTag,
		Content:    req.Content,
		CreatedAt:  time.Now(),
	}

	// Persist message
	h.db.Exec(
		`INSERT INTO chat_messages (id, channel, sender_id, sender_name, sender_tag, content, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.Channel, msg.SenderID, msg.SenderName, msg.SenderTag, msg.Content, msg.CreatedAt,
	)

	msgJSON := gin.H{
		"id":         msg.ID,
		"channel":    msg.Channel,
		"senderId":   msg.SenderID,
		"senderName": msg.SenderName,
		"senderTag":  msg.SenderTag,
		"content":    msg.Content,
		"timestamp":  msg.CreatedAt.UnixMilli(),
	}

	// Broadcast or send to alliance members
	if req.Channel == "world" {
		h.wsHub.Broadcast("chat:message", msgJSON)
	} else {
		// Alliance channel — send to all alliance members
		var allianceID string
		h.db.QueryRow("SELECT alliance_id FROM alliance_members WHERE player_id = ?", playerID).Scan(&allianceID)

		rows, err := h.db.Query("SELECT player_id FROM alliance_members WHERE alliance_id = ?", allianceID)
		if err == nil {
			defer rows.Close()
			var ids []string
			for rows.Next() {
				var pid string
				if err := rows.Scan(&pid); err == nil {
					ids = append(ids, pid)
				}
			}
			if len(ids) > 0 {
				h.wsHub.SendToMultiple(ids, "chat:message", msgJSON)
			}
		}
	}

	c.JSON(http.StatusCreated, msgJSON)
}

// History returns chat history for a channel.
func (h *ChatHandler) History(c *gin.Context) {
	channel := c.Query("channel")
	if channel != "world" && channel != "alliance" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel"})
		return
	}

	// For alliance channel, verify membership
	playerID := c.GetString("user_id")
	if channel == "alliance" {
		var allianceID string
		err := h.db.QueryRow("SELECT alliance_id FROM alliance_members WHERE player_id = ?", playerID).Scan(&allianceID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "not in an alliance"})
			return
		}
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	before := c.Query("before") // timestamp in ms

	var rows *sql.Rows
	var err error
	if before != "" {
		// Convert ms timestamp to time.Time
		var beforeTime time.Time
		if ts := parseTimestamp(before); ts != nil {
			beforeTime = *ts
			rows, err = h.db.Query(
				`SELECT id, channel, sender_id, sender_name, COALESCE(sender_tag, ''), content, created_at
				 FROM chat_messages WHERE channel = ? AND created_at < ?
				 ORDER BY created_at DESC LIMIT ?`,
				channel, beforeTime, limit,
			)
		} else {
			rows, err = h.db.Query(
				`SELECT id, channel, sender_id, sender_name, COALESCE(sender_tag, ''), content, created_at
				 FROM chat_messages WHERE channel = ?
				 ORDER BY created_at DESC LIMIT ?`,
				channel, limit,
			)
		}
	} else {
		rows, err = h.db.Query(
			`SELECT id, channel, sender_id, sender_name, COALESCE(sender_tag, ''), content, created_at
			 FROM chat_messages WHERE channel = ?
			 ORDER BY created_at DESC LIMIT ?`,
			channel, limit,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load messages"})
		return
	}
	defer rows.Close()

	var messages []gin.H
	for rows.Next() {
		var id, ch, senderID, senderName, senderTag, content string
		var createdAt time.Time
		if err := rows.Scan(&id, &ch, &senderID, &senderName, &senderTag, &content, &createdAt); err != nil {
			continue
		}
		messages = append(messages, gin.H{
			"id":         id,
			"channel":    ch,
			"senderId":   senderID,
			"senderName": senderName,
			"senderTag":  senderTag,
			"content":    content,
			"timestamp":  createdAt.UnixMilli(),
		})
	}

	if messages == nil {
		messages = []gin.H{}
	}
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// UpdateDND updates the player's chat DND mode.
func (h *ChatHandler) UpdateDND(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		Mode string `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validModes := map[string]bool{
		"none": true, "mute_world": true, "mute_alliance": true, "mute_all": true,
	}
	if !validModes[req.Mode] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid DND mode"})
		return
	}

	// Store DND mode in player_game_states as a JSON field
	var stateData string
	err := h.db.QueryRow(
		"SELECT state_data FROM player_game_states WHERE player_id = ?", playerID,
	).Scan(&stateData)
	if err != nil {
		// No state yet — create minimal state with DND
		state := map[string]interface{}{"chatDNDMode": req.Mode}
		patched, _ := json.Marshal(state)
		stateData = string(patched)
		h.db.Exec(
			"INSERT INTO player_game_states (player_id, state_data, updated_at) VALUES (?, ?, ?)",
			playerID, stateData, time.Now(),
		)
	} else {
		// Parse existing state, update DND field, marshal back
		var state map[string]interface{}
		if err := json.Unmarshal([]byte(stateData), &state); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "corrupt state data"})
			return
		}
		state["chatDNDMode"] = req.Mode
		patched, _ := json.Marshal(state)
		stateData = string(patched)
		h.db.Exec(
			"UPDATE player_game_states SET state_data = ?, updated_at = ? WHERE player_id = ?",
			stateData, time.Now(), playerID,
		)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "mode": req.Mode})
}

// --- Helpers ---

func parseTimestamp(s string) *time.Time {
	var ms int64
	if _, err := fmt.Sscanf(s, "%d", &ms); err != nil {
		return nil
	}
	t := time.UnixMilli(ms)
	return &t
}

