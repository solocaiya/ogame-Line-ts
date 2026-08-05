package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ogame-server/engine"
	"ogame-server/gamestate"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GameHandler handles game-related API endpoints.
type GameHandler struct {
	gameState *gamestate.GameState
	db        *sql.DB
}

// NewGameHandler creates a new GameHandler.
func NewGameHandler(gameState *gamestate.GameState, db *sql.DB) *GameHandler {
	return &GameHandler{gameState: gameState, db: db}
}

// GetGameState returns the current game state for the authenticated player.
// If not in memory, tries to load from database.
func (h *GameHandler) GetGameState(c *gin.Context) {
	playerID := c.GetString("user_id")
	playerJSON, ok := h.gameState.MarshalPlayer(playerID)
	if !ok && h.db != nil {
		// Try loading from database
		if err := h.gameState.LoadPlayer(h.db, playerID); err == nil {
			playerJSON, ok = h.gameState.MarshalPlayer(playerID)
		}
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "game state not found"})
		return
	}

	// Offline settlement: advance player state to current time
	// This processes fleet missions, queues, and resources that happened while offline
	now := time.Now().UnixMilli()
	_ = h.gameState.Advance(playerID, now)

	// Re-marshal after advance
	playerJSON, _ = h.gameState.MarshalPlayer(playerID)

	// Client expects { player: {...} }
	c.JSON(http.StatusOK, gin.H{"player": playerJSON})
}

// InitPlayer initializes a new player with a starting planet if not exists.
// If player has saved state in DB, loads it first.
func (h *GameHandler) InitPlayer(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		GameSpeed int `json:"gameSpeed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.GameSpeed = 1
	}
	if req.GameSpeed <= 0 {
		req.GameSpeed = 1
	}

	h.gameState.EnsurePlayer(playerID, req.GameSpeed)

	// Try loading from database if not in memory or no planets
	player, _ := h.gameState.GetPlayer(playerID)
	if len(player.Planets) == 0 && h.db != nil {
		_ = h.gameState.LoadPlayer(h.db, playerID)
		player, _ = h.gameState.GetPlayer(playerID)
	}

	if len(player.Planets) > 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": player})
		return
	}

	// Create starting planet [1:1:4]
	now := time.Now().UnixMilli()
	planet := &engine.PlanetState{
		ID:   "1-1-4",
		Name: "Homeworld",
		Coordinate: engine.Coordinate{Galaxy: 1, System: 1, Position: 4},
		Buildings:     map[string]int{"metalMine": 1, "crystalMine": 1, "deuteriumSynthesizer": 1, "solarPlant": 1},
		Technologies:  map[string]int{},
		Ships:         map[string]int{"smallCargo": 3, "espionageProbe": 1},
		Defenses:      map[string]int{"rocketLauncher": 5},
		Resources:     engine.Resources{Metal: 500, Crystal: 300, Deuterium: 100},
		StorageCap:    engine.Resources{Metal: 5000, Crystal: 5000, Deuterium: 5000, DarkMatter: 5000},
		Production:    engine.Resources{},
		BuildingQueue: []engine.BuildingQueueItem{},
		ResearchQueue: []engine.ResearchQueueItem{},
		ShipQueue:     []engine.ShipQueueItem{},
		DefenseQueue:  []engine.DefenseQueueItem{},
		LastUpdate:    now,
	}

	_ = h.gameState.UpdatePlayer(playerID, func(p *engine.PlayerState) error {
		p.Planets[planet.ID] = planet
		return nil
	})

	playerJSON, _ := h.gameState.MarshalPlayer(playerID)
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": playerJSON})
}

// StartBuilding starts a building construction.
func (h *GameHandler) StartBuilding(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		PlanetID     string `json:"planetId" binding:"required"`
		BuildingType string `json:"buildingType" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	buildingDef, ok := engine.BuildingDefs[req.BuildingType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid building type"})
		return
	}

	var endTime int64
	err := h.gameState.UpdatePlanet(playerID, req.PlanetID, func(planet *engine.PlanetState) error {
		// Check queue not full (max 1 building at a time)
		if len(planet.BuildingQueue) > 0 {
			return fmt.Errorf("building queue is full")
		}

		currentLevel := planet.Buildings[req.BuildingType]
		targetLevel := currentLevel + 1
		cost := engine.CalculateBuildingCost(buildingDef.BaseCost, buildingDef.CostMultiplier, targetLevel)

		costRes := engine.Resources{Metal: cost.Metal, Crystal: cost.Crystal, Deuterium: cost.Deuterium}
		if !planet.Resources.CanAfford(costRes) {
			return fmt.Errorf("insufficient resources")
		}

		roboticsLevel := planet.Buildings["roboticsFactory"]
		naniteLevel := planet.Buildings["naniteFactory"]
		buildTime := engine.CalculateBuildingTime(buildingDef.BaseTime, roboticsLevel, naniteLevel, 0)

		planet.Resources = planet.Resources.Sub(costRes)
		endTime = time.Now().UnixMilli() + int64(buildTime*1000)
		planet.BuildingQueue = append(planet.BuildingQueue, engine.BuildingQueueItem{
			Type:        req.BuildingType,
			TargetLevel: targetLevel,
			EndTime:     endTime,
		})
		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "endTime": endTime})
}

// StartShipProduction starts ship production.
func (h *GameHandler) StartShipProduction(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		PlanetID string `json:"planetId" binding:"required"`
		ShipType string `json:"shipType" binding:"required"`
		Count    int    `json:"count" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shipDef, ok := engine.ShipDefs[req.ShipType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ship type"})
		return
	}

	var endTime int64
	err := h.gameState.UpdatePlanet(playerID, req.PlanetID, func(planet *engine.PlanetState) error {
		totalCost := engine.Resources{
			Metal:     shipDef.Cost.Metal * int64(req.Count),
			Crystal:   shipDef.Cost.Crystal * int64(req.Count),
			Deuterium: shipDef.Cost.Deuterium * int64(req.Count),
		}

		if !planet.Resources.CanAfford(totalCost) {
			return fmt.Errorf("insufficient resources")
		}

		shipyardLevel := planet.Buildings["shipyard"]
		naniteLevel := planet.Buildings["naniteFactory"]
		buildTime := engine.CalculateShipBuildTime(shipDef.BuildTime, shipyardLevel, naniteLevel) * req.Count

		planet.Resources = planet.Resources.Sub(totalCost)
		endTime = time.Now().UnixMilli() + int64(buildTime*1000)
		planet.ShipQueue = append(planet.ShipQueue, engine.ShipQueueItem{
			Type:    req.ShipType,
			Count:   req.Count,
			EndTime: endTime,
		})
		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "endTime": endTime})
}

// StartDefenseProduction starts defense production.
func (h *GameHandler) StartDefenseProduction(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		PlanetID    string `json:"planetId" binding:"required"`
		DefenseType string `json:"defenseType" binding:"required"`
		Count       int    `json:"count" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	defenseDef, ok := engine.DefenseDefs[req.DefenseType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid defense type"})
		return
	}

	var endTime int64
	err := h.gameState.UpdatePlanet(playerID, req.PlanetID, func(planet *engine.PlanetState) error {
		totalCost := engine.Resources{
			Metal:     defenseDef.Cost.Metal * int64(req.Count),
			Crystal:   defenseDef.Cost.Crystal * int64(req.Count),
			Deuterium: defenseDef.Cost.Deuterium * int64(req.Count),
		}

		if !planet.Resources.CanAfford(totalCost) {
			return fmt.Errorf("insufficient resources")
		}

		shipyardLevel := planet.Buildings["shipyard"]
		naniteLevel := planet.Buildings["naniteFactory"]
		buildTime := engine.CalculateDefenseBuildTime(defenseDef.BuildTime, shipyardLevel, naniteLevel) * req.Count

		planet.Resources = planet.Resources.Sub(totalCost)
		endTime = time.Now().UnixMilli() + int64(buildTime*1000)
		planet.DefenseQueue = append(planet.DefenseQueue, engine.DefenseQueueItem{
			Type:    req.DefenseType,
			Count:   req.Count,
			EndTime: endTime,
		})
		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "endTime": endTime})
}

// validMissionTypes is the set of allowed fleet mission types.
var validMissionTypes = map[string]bool{
	"attack":     true,
	"deploy":     true,
	"transport":  true,
	"spy":        true,
	"recycle":    true,
	"colonize":   true,
	"expedition": true,
}

// SendFleet dispatches a fleet mission.
func (h *GameHandler) SendFleet(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		PlanetID       string            `json:"planetId" binding:"required"`
		TargetCoord    engine.Coordinate `json:"targetCoordinate" binding:"required"`
		MissionType    string            `json:"missionType" binding:"required"`
		Fleet          map[string]int    `json:"fleet" binding:"required"`
		Cargo          engine.Resources  `json:"cargo"`
		BattleToFinish bool              `json:"battleToFinish"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !validMissionTypes[req.MissionType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mission type"})
		return
	}

	var mission engine.FleetMission
	err := h.gameState.UpdatePlayer(playerID, func(p *engine.PlayerState) error {
		planet, ok := p.Planets[req.PlanetID]
		if !ok {
			return fmt.Errorf("planet not found: %s", req.PlanetID)
		}

		// Check if player has the ships
		for shipType, count := range req.Fleet {
			if planet.Ships[shipType] < count {
				return fmt.Errorf("insufficient ships: %s", shipType)
			}
		}

		distance := engine.CalculateDistance(planet.Coordinate, req.TargetCoord)
		minSpeed := engine.CalculateMinSpeed(req.Fleet)
		flightTime := engine.CalculateFlightTime(distance, minSpeed)
		fuel := engine.CalculateFuelConsumption(req.Fleet, distance, flightTime)

		if planet.Resources.Deuterium < fuel {
			return fmt.Errorf("insufficient fuel: need %d, have %d", fuel, planet.Resources.Deuterium)
		}

		// Check cargo affordability BEFORE any deductions (atomic validation)
		totalDeuteriumNeeded := fuel + req.Cargo.Deuterium
		if req.Cargo.Metal > 0 || req.Cargo.Crystal > 0 || req.Cargo.Deuterium > 0 {
			// Check against resources after fuel deduction
			availableAfterFuel := planet.Resources
			availableAfterFuel.Deuterium -= fuel
			if !availableAfterFuel.CanAfford(req.Cargo) {
				return fmt.Errorf("insufficient cargo resources")
			}
		}

		// All checks passed — now deduct everything atomically
		for shipType, count := range req.Fleet {
			planet.Ships[shipType] -= count
		}
		planet.Resources.Deuterium -= fuel
		if req.Cargo.Metal > 0 || req.Cargo.Crystal > 0 || req.Cargo.Deuterium > 0 {
			planet.Resources = planet.Resources.Sub(req.Cargo)
		}

		now := time.Now().UnixMilli()
		mission = engine.FleetMission{
			ID:             generateID(),
			PlayerID:       playerID,
			MissionType:    req.MissionType,
			Fleet:          req.Fleet,
			Cargo:          req.Cargo,
			Origin:         planet.Coordinate,
			Target:         req.TargetCoord,
			DepartureTime:  now,
			ArrivalTime:    now + int64(flightTime*1000),
			ReturnTime:     now + int64(flightTime*2000),
			Status:         "outbound",
			BattleToFinish: req.BattleToFinish,
		}

		// Add mission to player — same lock, atomic with ship/fuel deduction
		p.FleetMissions = append(p.FleetMissions, mission)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"mission": mission,
	})
}

// RecallFleet recalls an in-flight fleet mission.
// The fleet will reverse course and return to origin.
func (h *GameHandler) RecallFleet(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		MissionID string `json:"missionId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mission, err := h.gameState.RecallFleet(playerID, req.MissionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"mission": mission,
	})
}

// CancelBuilding cancels a building construction.
func (h *GameHandler) CancelBuilding(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		PlanetID string `json:"planetId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.gameState.UpdatePlanet(playerID, req.PlanetID, func(planet *engine.PlanetState) error {
		if len(planet.BuildingQueue) == 0 {
			return fmt.Errorf("no building in queue")
		}

		item := planet.BuildingQueue[0]
		buildingDef := engine.BuildingDefs[item.Type]
		cost := engine.CalculateBuildingCost(buildingDef.BaseCost, buildingDef.CostMultiplier, item.TargetLevel)
		refund := engine.Resources{
			Metal:     cost.Metal / 2,
			Crystal:   cost.Crystal / 2,
			Deuterium: cost.Deuterium / 2,
		}
		planet.Resources = planet.Resources.Add(refund)
		planet.BuildingQueue = planet.BuildingQueue[1:]
		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// CancelShipProduction cancels ship production.
func (h *GameHandler) CancelShipProduction(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		PlanetID string `json:"planetId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.gameState.UpdatePlanet(playerID, req.PlanetID, func(planet *engine.PlanetState) error {
		if len(planet.ShipQueue) == 0 {
			return fmt.Errorf("no ship in queue")
		}

		item := planet.ShipQueue[0]
		shipDef := engine.ShipDefs[item.Type]
		refund := engine.Resources{
			Metal:     shipDef.Cost.Metal * int64(item.Count) / 2,
			Crystal:   shipDef.Cost.Crystal * int64(item.Count) / 2,
			Deuterium: shipDef.Cost.Deuterium * int64(item.Count) / 2,
		}
		planet.Resources = planet.Resources.Add(refund)
		planet.ShipQueue = planet.ShipQueue[1:]
		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// CancelDefenseProduction cancels defense production.
func (h *GameHandler) CancelDefenseProduction(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		PlanetID string `json:"planetId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.gameState.UpdatePlanet(playerID, req.PlanetID, func(planet *engine.PlanetState) error {
		if len(planet.DefenseQueue) == 0 {
			return fmt.Errorf("no defense in queue")
		}

		item := planet.DefenseQueue[0]
		defenseDef := engine.DefenseDefs[item.Type]
		refund := engine.Resources{
			Metal:     defenseDef.Cost.Metal * int64(item.Count) / 2,
			Crystal:   defenseDef.Cost.Crystal * int64(item.Count) / 2,
			Deuterium: defenseDef.Cost.Deuterium * int64(item.Count) / 2,
		}
		planet.Resources = planet.Resources.Add(refund)
		planet.DefenseQueue = planet.DefenseQueue[1:]
		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// StartResearch starts a technology research.
func (h *GameHandler) StartResearch(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		PlanetID      string `json:"planetId" binding:"required"`
		ResearchType  string `json:"researchType" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Look up research definition from config
	researchDef, ok := engine.ResearchDefs[req.ResearchType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid research type"})
		return
	}

	var endTime int64
	err := h.gameState.UpdatePlanet(playerID, req.PlanetID, func(planet *engine.PlanetState) error {
		// Check research lab exists
		labLevel := planet.Buildings["researchLab"]
		if labLevel <= 0 {
			return fmt.Errorf("research lab required")
		}

		// Check queue not full (max 1 research at a time)
		if len(planet.ResearchQueue) > 0 {
			return fmt.Errorf("research queue is full")
		}

		currentLevel := planet.Technologies[req.ResearchType]
		targetLevel := currentLevel + 1

		// Exponential cost scaling: floor(baseCost * multiplier^(level-1))
		cost := engine.CalculateResearchCost(researchDef.BaseCost, researchDef.CostMultiplier, targetLevel)
		totalCost := engine.Resources{
			Metal:     cost.Metal,
			Crystal:   cost.Crystal,
			Deuterium: cost.Deuterium,
		}

		if !planet.Resources.CanAfford(totalCost) {
			return fmt.Errorf("insufficient resources")
		}

		// Research time = baseTime * level / (1 + labLevel * 0.1)
		baseTime := researchDef.BaseTime * targetLevel
		buildTime := int(float64(baseTime) / (1.0 + float64(labLevel)*0.1))
		if buildTime < 1 {
			buildTime = 1
		}

		planet.Resources = planet.Resources.Sub(totalCost)
		endTime = time.Now().UnixMilli() + int64(buildTime*1000)
		planet.ResearchQueue = append(planet.ResearchQueue, engine.ResearchQueueItem{
			Type:    req.ResearchType,
			EndTime: endTime,
		})
		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "endTime": endTime})
}

// CancelResearch cancels a research.
func (h *GameHandler) CancelResearch(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		PlanetID string `json:"planetId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.gameState.UpdatePlanet(playerID, req.PlanetID, func(planet *engine.PlanetState) error {
		if len(planet.ResearchQueue) == 0 {
			return fmt.Errorf("no research in queue")
		}
		// Just remove the queue item, no refund for research
		planet.ResearchQueue = planet.ResearchQueue[1:]
		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// generateID generates a unique ID using UUID.
func generateID() string {
	return uuid.New().String()
}

// GetLeaderboard handles GET /api/game/leaderboard
func (h *GameHandler) GetLeaderboard(c *gin.Context) {
	playerID := c.GetString("user_id")
	if playerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	entries, err := gamestate.GetLeaderboard(h.db, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}
	if entries == nil {
		entries = []gamestate.LeaderboardEntry{}
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"count":   len(entries),
		"limit":   limit,
		"offset":  offset,
	})
}

// GetNotifications handles GET /api/game/notifications
func (h *GameHandler) GetNotifications(c *gin.Context) {
	playerID := c.GetString("user_id")
	if playerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	unreadOnly := c.Query("unread_only") == "true"

	query := `SELECT id, type, message, data, created_at, read FROM notifications WHERE player_id = ?`
	args := []interface{}{playerID}
	if unreadOnly {
		query += ` AND read = 0`
	}
	query += ` ORDER BY created_at DESC LIMIT ?`
	args = append(args, limit)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}
	defer rows.Close()

	type notifResponse struct {
		ID        string      `json:"id"`
		Type      string      `json:"type"`
		Message   string      `json:"message"`
		Data      interface{} `json:"data,omitempty"`
		CreatedAt int64       `json:"createdAt"`
		Read      bool        `json:"read"`
	}

	notifications := []notifResponse{}
	for rows.Next() {
		var n notifResponse
		var dataStr string
		var createdAt int64
		if err := rows.Scan(&n.ID, &n.Type, &n.Message, &dataStr, &createdAt, &n.Read); err != nil {
			continue
		}
		if dataStr != "" {
			_ = json.Unmarshal([]byte(dataStr), &n.Data)
		}
		n.CreatedAt = createdAt
		notifications = append(notifications, n)
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

// MarkNotificationRead handles POST /api/game/notifications/:id/read
func (h *GameHandler) MarkNotificationRead(c *gin.Context) {
	playerID := c.GetString("user_id")
	if playerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notifID := c.Param("id")
	if notifID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification id required"})
		return
	}

	result, err := h.db.Exec(`UPDATE notifications SET read = 1 WHERE id = ? AND player_id = ?`, notifID, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification read"})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetBattleReplays handles GET /api/game/battle-replays
func (h *GameHandler) GetBattleReplays(c *gin.Context) {
	playerID := c.GetString("user_id")
	if playerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	rows, err := h.db.Query(`
		SELECT id, attacker_id, target_coord, result_data, created_at
		FROM battle_replays
		WHERE attacker_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, playerID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch battle replays"})
		return
	}
	defer rows.Close()

	type replayResponse struct {
		ID          string          `json:"id"`
		AttackerID  string          `json:"attackerId"`
		TargetCoord string          `json:"targetCoord"`
		Result      json.RawMessage `json:"result"`
		CreatedAt   string          `json:"createdAt"`
	}

	replays := []replayResponse{}
	for rows.Next() {
		var r replayResponse
		var resultData string
		if err := rows.Scan(&r.ID, &r.AttackerID, &r.TargetCoord, &resultData, &r.CreatedAt); err != nil {
			continue
		}
		r.Result = json.RawMessage(resultData)
		replays = append(replays, r)
	}

	c.JSON(http.StatusOK, gin.H{
		"replays": replays,
		"count":   len(replays),
	})
}

// UpdateSettings updates the player's game settings (e.g. battle mode toggle).
func (h *GameHandler) UpdateSettings(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		BattleToFinish bool `json:"battleToFinish"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	settings := engine.PlayerSettings{
		BattleToFinish: req.BattleToFinish,
	}

	if err := h.gameState.UpdateSettings(playerID, settings); err != nil {
		// Player might not be in memory yet — try loading from DB first
		if h.db != nil {
			if loadErr := h.gameState.LoadPlayer(h.db, playerID); loadErr == nil {
				if updateErr := h.gameState.UpdateSettings(playerID, settings); updateErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "settings updated", "settings": settings})
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "settings updated", "settings": settings})
}

// GetSettings returns the player's current game settings.
func (h *GameHandler) GetSettings(c *gin.Context) {
	playerID := c.GetString("user_id")

	settings, ok := h.gameState.GetSettings(playerID)
	if !ok && h.db != nil {
		if err := h.gameState.LoadPlayer(h.db, playerID); err == nil {
			settings, ok = h.gameState.GetSettings(playerID)
		}
	}
	if !ok {
		// Return defaults if player not found
		c.JSON(http.StatusOK, gin.H{"settings": engine.PlayerSettings{BattleToFinish: false}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}
