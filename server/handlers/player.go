package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"ogame-server/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SaveGameRequest struct {
	SaveSlot     string `json:"slot"`
	GameData     string `json:"gameData" binding:"required"`
	NPCData      string `json:"npcData"`
	UniverseData string `json:"universeData"`
}

type PlayerHandler struct{}

func (h *PlayerHandler) SaveGame(c *gin.Context) {
	userID := c.GetString("user_id")

	var req SaveGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if req.SaveSlot == "" {
		req.SaveSlot = "default"
	}

	saveID := uuid.New().String()
	now := time.Now()

	// Upsert: insert or replace existing save for this user+slot
	_, err := database.DB.Exec(
		`INSERT INTO player_saves (id, user_id, save_slot, game_data, npc_data, universe_data, saved_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(user_id, save_slot) DO UPDATE SET
		   game_data = excluded.game_data,
		   npc_data = excluded.npc_data,
		   universe_data = excluded.universe_data,
		   saved_at = excluded.saved_at`,
		saveID, userID, req.SaveSlot, req.GameData, req.NPCData, req.UniverseData, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save game: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "game saved successfully",
		"id":        saveID,
		"slot":      req.SaveSlot,
		"savedAt":   now,
	})
}

func (h *PlayerHandler) LoadGame(c *gin.Context) {
	userID := c.GetString("user_id")
	saveSlot := c.DefaultQuery("slot", "default")

	var gameData, npcData, universeData string
	var savedAt time.Time

	err := database.DB.QueryRow(
		`SELECT game_data, npc_data, universe_data, saved_at
		 FROM player_saves WHERE user_id = ? AND save_slot = ?`,
		userID, saveSlot,
	).Scan(&gameData, &npcData, &universeData, &savedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "no save found for this slot"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load game: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"gameData":     gameData,
		"npcData":      npcData,
		"universeData": universeData,
		"savedAt":      savedAt,
	})
}

func (h *PlayerHandler) ListSaves(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := database.DB.Query(
		`SELECT save_slot, saved_at FROM player_saves WHERE user_id = ? ORDER BY saved_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list saves: " + err.Error()})
		return
	}
	defer rows.Close()

	type SaveInfo struct {
		SaveSlot string    `json:"slot"`
		SavedAt  time.Time `json:"savedAt"`
	}
	var saves []SaveInfo
	for rows.Next() {
		var s SaveInfo
		if err := rows.Scan(&s.SaveSlot, &s.SavedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read save info"})
			return
		}
		saves = append(saves, s)
	}

	if saves == nil {
		saves = []SaveInfo{}
	}

	c.JSON(http.StatusOK, saves)
}

func (h *PlayerHandler) DeleteSave(c *gin.Context) {
	userID := c.GetString("user_id")
	saveSlot := c.DefaultQuery("slot", "default")

	result, err := database.DB.Exec(
		`DELETE FROM player_saves WHERE user_id = ? AND save_slot = ?`,
		userID, saveSlot,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete save: " + err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no save found for this slot"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "save deleted successfully"})
}
