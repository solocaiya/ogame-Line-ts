package handlers

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"time"

	"ogame-server/engine"
	"ogame-server/gamestate"
	"ogame-server/ws"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AccelerateHandler handles acceleration endpoints for building, research,
// ship construction, and fleet travel.
type AccelerateHandler struct {
	gameState *gamestate.GameState
	db        *sql.DB
	wsHub     *ws.Hub
}

// NewAccelerateHandler creates a new AccelerateHandler.
func NewAccelerateHandler(gameState *gamestate.GameState, db *sql.DB, wsHub *ws.Hub) *AccelerateHandler {
	return &AccelerateHandler{gameState: gameState, db: db, wsHub: wsHub}
}

// ─── Cost Calculation ────────────────────────────────────────────────────────

// calcAccelerateCost returns the DM cost to skip the given number of milliseconds.
// 1 DM = 60 minutes (AccelerateCostPerHour). Fractional hours round up.
func calcAccelerateCost(remainingMs int64) int64 {
	if remainingMs <= 0 {
		return 0
	}
	remainingMinutes := float64(remainingMs) / 60000.0
	hours := math.Ceil(remainingMinutes / 60.0)
	if hours < 1 {
		hours = 1
	}
	return int64(hours) * engine.AccelerateCostPerHour
}

// deductDM deducts the given amount from the player's dark matter balance and
// logs the transaction. Returns the new balance. Caller must hold the DB tx.
func (h *AccelerateHandler) deductDM(tx *sql.Tx, playerID string, amount int64, refType, refID string) (int64, error) {
	// Check balance
	var balance int64
	err := tx.QueryRow(`SELECT dark_matter_balance FROM users WHERE id = ?`, playerID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to read balance: %w", err)
	}
	if balance < amount {
		return 0, fmt.Errorf("insufficient dark matter")
	}

	newBalance := balance - amount
	_, err = tx.Exec(`UPDATE users SET dark_matter_balance = ?, consumption_points = consumption_points + ? WHERE id = ?`,
		newBalance, amount, playerID)
	if err != nil {
		return 0, fmt.Errorf("failed to update balance: %w", err)
	}

	// Log transaction
	txID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = tx.Exec(`
		INSERT INTO dark_matter_transactions (id, user_id, amount, balance_after, type, ref_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, txID, playerID, -amount, newBalance, refType, refID, now)
	if err != nil {
		return 0, fmt.Errorf("failed to log transaction: %w", err)
	}

	return newBalance, nil
}

// ─── Building Acceleration ────────────────────────────────────────────────────

type accelerateBuildingRequest struct {
	PlanetID    string `json:"planetId" binding:"required"`
	QueueIndex  int    `json:"queueIndex"`
	SkipMinutes int64  `json:"skipMinutes" binding:"required"`
}

// AccelerateBuilding accelerates a building construction queue item.
func (h *AccelerateHandler) AccelerateBuilding(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req accelerateBuildingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	planet, ok := h.gameState.GetPlanet(playerID, req.PlanetID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "planet not found"})
		return
	}

	if req.QueueIndex < 0 || req.QueueIndex >= len(planet.BuildingQueue) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid queue index"})
		return
	}

	now := time.Now().UnixMilli()
	item := planet.BuildingQueue[req.QueueIndex]
	remainingMs := item.EndTime - now
	if remainingMs <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "queue item already complete"})
		return
	}

	skipMs := req.SkipMinutes * 60 * 1000
	if skipMs > remainingMs {
		skipMs = remainingMs
	}

	costDM := calcAccelerateCost(skipMs)
	if costDM <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nothing to accelerate"})
		return
	}

	// Deduct DM
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	newBalance, err := h.deductDM(tx, playerID, costDM, "accelerate", fmt.Sprintf("building:%s:%d", req.PlanetID, req.QueueIndex))
	if err != nil {
		if err.Error() == "insufficient dark matter" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient dark matter", "costDM": costDM})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	// Update queue item EndTime
	newEndTime := item.EndTime - skipMs
	err = h.gameState.UpdatePlanet(playerID, req.PlanetID, func(planet *engine.PlanetState) error {
		if req.QueueIndex >= len(planet.BuildingQueue) {
			return fmt.Errorf("queue index out of range")
		}
		planet.BuildingQueue[req.QueueIndex].EndTime = newEndTime
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update queue"})
		return
	}

	// Notify
	if h.wsHub != nil {
		h.wsHub.SendTo(playerID, "dmBalanceChanged", gin.H{
			"balance":  newBalance,
			"planetId": req.PlanetID,
		})
		h.wsHub.SendTo(playerID, "buildingQueueUpdated", gin.H{
			"planetId":   req.PlanetID,
			"queueIndex": req.QueueIndex,
			"newEndTime": newEndTime,
			"skippedMs":  skipMs,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"costDM":     costDM,
		"newEndTime": newEndTime,
		"skippedMs":  skipMs,
		"newBalance": newBalance,
	})
}

// ─── Research Acceleration ────────────────────────────────────────────────────

type accelerateResearchRequest struct {
	PlanetID    string `json:"planetId" binding:"required"`
	QueueIndex  int    `json:"queueIndex"`
	SkipMinutes int64  `json:"skipMinutes" binding:"required"`
}

// AccelerateResearch accelerates a research queue item.
func (h *AccelerateHandler) AccelerateResearch(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req accelerateResearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	planet, ok := h.gameState.GetPlanet(playerID, req.PlanetID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "planet not found"})
		return
	}

	if req.QueueIndex < 0 || req.QueueIndex >= len(planet.ResearchQueue) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid queue index"})
		return
	}

	now := time.Now().UnixMilli()
	item := planet.ResearchQueue[req.QueueIndex]
	remainingMs := item.EndTime - now
	if remainingMs <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "queue item already complete"})
		return
	}

	skipMs := req.SkipMinutes * 60 * 1000
	if skipMs > remainingMs {
		skipMs = remainingMs
	}

	costDM := calcAccelerateCost(skipMs)
	if costDM <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nothing to accelerate"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	newBalance, err := h.deductDM(tx, playerID, costDM, "accelerate", fmt.Sprintf("research:%s:%d", req.PlanetID, req.QueueIndex))
	if err != nil {
		if err.Error() == "insufficient dark matter" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient dark matter", "costDM": costDM})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	newEndTime := item.EndTime - skipMs
	err = h.gameState.UpdatePlanet(playerID, req.PlanetID, func(planet *engine.PlanetState) error {
		if req.QueueIndex >= len(planet.ResearchQueue) {
			return fmt.Errorf("queue index out of range")
		}
		planet.ResearchQueue[req.QueueIndex].EndTime = newEndTime
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update queue"})
		return
	}

	if h.wsHub != nil {
		h.wsHub.SendTo(playerID, "dmBalanceChanged", gin.H{
			"balance":  newBalance,
			"planetId": req.PlanetID,
		})
		h.wsHub.SendTo(playerID, "researchQueueUpdated", gin.H{
			"planetId":   req.PlanetID,
			"queueIndex": req.QueueIndex,
			"newEndTime": newEndTime,
			"skippedMs":  skipMs,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"costDM":     costDM,
		"newEndTime": newEndTime,
		"skippedMs":  skipMs,
		"newBalance": newBalance,
	})
}

// ─── Ship Build Acceleration ─────────────────────────────────────────────────

type accelerateFleetBuildRequest struct {
	PlanetID    string `json:"planetId" binding:"required"`
	QueueIndex  int    `json:"queueIndex"`
	SkipMinutes int64  `json:"skipMinutes" binding:"required"`
}

// AccelerateFleetBuild accelerates a ship construction queue item.
func (h *AccelerateHandler) AccelerateFleetBuild(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req accelerateFleetBuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	planet, ok := h.gameState.GetPlanet(playerID, req.PlanetID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "planet not found"})
		return
	}

	if req.QueueIndex < 0 || req.QueueIndex >= len(planet.ShipQueue) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid queue index"})
		return
	}

	now := time.Now().UnixMilli()
	item := planet.ShipQueue[req.QueueIndex]
	remainingMs := item.EndTime - now
	if remainingMs <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "queue item already complete"})
		return
	}

	skipMs := req.SkipMinutes * 60 * 1000
	if skipMs > remainingMs {
		skipMs = remainingMs
	}

	costDM := calcAccelerateCost(skipMs)
	if costDM <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nothing to accelerate"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	newBalance, err := h.deductDM(tx, playerID, costDM, "accelerate", fmt.Sprintf("ship_build:%s:%d", req.PlanetID, req.QueueIndex))
	if err != nil {
		if err.Error() == "insufficient dark matter" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient dark matter", "costDM": costDM})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	newEndTime := item.EndTime - skipMs
	err = h.gameState.UpdatePlanet(playerID, req.PlanetID, func(planet *engine.PlanetState) error {
		if req.QueueIndex >= len(planet.ShipQueue) {
			return fmt.Errorf("queue index out of range")
		}
		planet.ShipQueue[req.QueueIndex].EndTime = newEndTime
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update queue"})
		return
	}

	if h.wsHub != nil {
		h.wsHub.SendTo(playerID, "dmBalanceChanged", gin.H{
			"balance":  newBalance,
			"planetId": req.PlanetID,
		})
		h.wsHub.SendTo(playerID, "shipQueueUpdated", gin.H{
			"planetId":   req.PlanetID,
			"queueIndex": req.QueueIndex,
			"newEndTime": newEndTime,
			"skippedMs":  skipMs,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"costDM":     costDM,
		"newEndTime": newEndTime,
		"skippedMs":  skipMs,
		"newBalance": newBalance,
	})
}

// ─── Fleet Travel Acceleration ───────────────────────────────────────────────

type accelerateFleetTravelRequest struct {
	MissionID   string `json:"missionId" binding:"required"`
	SkipMinutes int64  `json:"skipMinutes" binding:"required"`
}

// AccelerateFleetTravel accelerates a fleet mission's travel time.
func (h *AccelerateHandler) AccelerateFleetTravel(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req accelerateFleetTravelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	player, ok := h.gameState.GetPlayer(playerID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}

	// Find the mission
	missionIdx := -1
	for i, m := range player.FleetMissions {
		if m.ID == req.MissionID && m.PlayerID == playerID {
			missionIdx = i
			break
		}
	}
	if missionIdx < 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "fleet mission not found"})
		return
	}

	mission := player.FleetMissions[missionIdx]
	now := time.Now().UnixMilli()

	// Determine which arrival time to reduce based on status
	var targetTime *int64
	switch mission.Status {
	case "outbound":
		targetTime = &mission.ArrivalTime
	case "returning":
		targetTime = &mission.ReturnTime
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "mission not in transit"})
		return
	}

	remainingMs := *targetTime - now
	if remainingMs <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fleet already arrived"})
		return
	}

	skipMs := req.SkipMinutes * 60 * 1000
	if skipMs > remainingMs {
		skipMs = remainingMs
	}

	costDM := calcAccelerateCost(skipMs)
	if costDM <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nothing to accelerate"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	newBalance, err := h.deductDM(tx, playerID, costDM, "accelerate", fmt.Sprintf("fleet_travel:%s", req.MissionID))
	if err != nil {
		if err.Error() == "insufficient dark matter" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient dark matter", "costDM": costDM})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	// Update the fleet mission timing
	newTargetTime := *targetTime - skipMs
	err = h.gameState.UpdatePlayer(playerID, func(player *engine.PlayerState) error {
		for i, m := range player.FleetMissions {
			if m.ID == req.MissionID {
				switch m.Status {
				case "outbound":
					player.FleetMissions[i].ArrivalTime = newTargetTime
				case "returning":
					player.FleetMissions[i].ReturnTime = newTargetTime
				}
				return nil
			}
		}
		return fmt.Errorf("mission not found")
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update fleet mission"})
		return
	}

	if h.wsHub != nil {
		h.wsHub.SendTo(playerID, "dmBalanceChanged", gin.H{
			"balance": newBalance,
		})
		h.wsHub.SendTo(playerID, "fleetMissionUpdated", gin.H{
			"missionId":      req.MissionID,
			"newTargetTime":  newTargetTime,
			"skippedMs":      skipMs,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"costDM":        costDM,
		"newTargetTime": newTargetTime,
		"skippedMs":     skipMs,
		"newBalance":    newBalance,
	})
}

// ─── Available Accelerations ─────────────────────────────────────────────────

// GetAvailable returns all currently acceleratable queue items and fleet missions
// with their remaining time and DM cost.
func (h *AccelerateHandler) GetAvailable(c *gin.Context) {
	playerID := c.GetString("user_id")

	player, ok := h.gameState.GetPlayer(playerID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}

	now := time.Now().UnixMilli()
	type availableItem struct {
		Type        string `json:"type"`        // building, research, ship_build, fleet_travel
		PlanetID    string `json:"planetId"`    // empty for fleet_travel
		QueueIndex  int    `json:"queueIndex"`  // -1 for fleet_travel
		MissionID   string `json:"missionId"`   // only for fleet_travel
		ItemName    string `json:"itemName"`
		RemainingMs int64  `json:"remainingMs"`
		CostDM      int64  `json:"costDM"`
	}

	items := make([]availableItem, 0)

	for planetID, planet := range player.Planets {
		// Building queue
		for i, item := range planet.BuildingQueue {
			remaining := item.EndTime - now
			if remaining > 0 {
				items = append(items, availableItem{
					Type:        "building",
					PlanetID:    planetID,
					QueueIndex:  i,
					ItemName:    item.Type,
					RemainingMs: remaining,
					CostDM:      calcAccelerateCost(remaining),
				})
			}
		}

		// Research queue
		for i, item := range planet.ResearchQueue {
			remaining := item.EndTime - now
			if remaining > 0 {
				items = append(items, availableItem{
					Type:        "research",
					PlanetID:    planetID,
					QueueIndex:  i,
					ItemName:    item.Type,
					RemainingMs: remaining,
					CostDM:      calcAccelerateCost(remaining),
				})
			}
		}

		// Ship queue
		for i, item := range planet.ShipQueue {
			remaining := item.EndTime - now
			if remaining > 0 {
				items = append(items, availableItem{
					Type:        "ship_build",
					PlanetID:    planetID,
					QueueIndex:  i,
					ItemName:    item.Type,
					RemainingMs: remaining,
					CostDM:      calcAccelerateCost(remaining),
				})
			}
		}

		// Defense queue
		for i, item := range planet.DefenseQueue {
			remaining := item.EndTime - now
			if remaining > 0 {
				items = append(items, availableItem{
					Type:        "defense_build",
					PlanetID:    planetID,
					QueueIndex:  i,
					ItemName:    item.Type,
					RemainingMs: remaining,
					CostDM:      calcAccelerateCost(remaining),
				})
			}
		}
	}

	// Fleet missions in transit
	for _, m := range player.FleetMissions {
		var targetTime int64
		switch m.Status {
		case "outbound":
			targetTime = m.ArrivalTime
		case "returning":
			targetTime = m.ReturnTime
		default:
			continue
		}
		remaining := targetTime - now
		if remaining > 0 {
			items = append(items, availableItem{
				Type:        "fleet_travel",
				MissionID:   m.ID,
				ItemName:    m.MissionType,
				RemainingMs: remaining,
				CostDM:      calcAccelerateCost(remaining),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
	})
}
