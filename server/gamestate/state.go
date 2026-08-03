package gamestate

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"ogame-server/engine"
)

// GameEvent represents an event that should be broadcast to players.
type GameEvent struct {
	Type     string      `json:"type"`
	PlayerID string      `json:"playerId,omitempty"` // empty = broadcast to all
	Data     interface{} `json:"data"`
}

// GameState manages the authoritative server-side game state.
type GameState struct {
	mu           sync.RWMutex
	players      map[string]*engine.PlayerState // playerID -> state
	eventHandler atomic.Value                    // stores func(event GameEvent); read lock-free
}

// New creates a new GameState.
func New() *GameState {
	return &GameState{
		players: make(map[string]*engine.PlayerState),
	}
}

// SetEventHandler sets a callback for game events (e.g., broadcast to WebSocket).
func (gs *GameState) SetEventHandler(handler func(event GameEvent)) {
	gs.eventHandler.Store(handler)
}

func (gs *GameState) emitEvent(event GameEvent) {
	if v := gs.eventHandler.Load(); v != nil {
		v.(func(event GameEvent))(event)
	}
}

// GetPlayer returns a player's state (read-only copy via deep copy).
// For JSON responses, prefer MarshalPlayer to avoid redundant serialization.
func (gs *GameState) GetPlayer(playerID string) (*engine.PlayerState, bool) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	p, ok := gs.players[playerID]
	if !ok {
		return nil, false
	}
	data, _ := json.Marshal(p)
	var copy engine.PlayerState
	json.Unmarshal(data, &copy)
	return &copy, true
}

// MarshalPlayer serializes a player's state directly under the read lock,
// avoiding the Marshal→Unmarshal→Marshal triple-serialization of GetPlayer.
// Returns json.RawMessage ready for embedding in a response.
func (gs *GameState) MarshalPlayer(playerID string) (json.RawMessage, bool) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	p, ok := gs.players[playerID]
	if !ok {
		return nil, false
	}
	data, err := json.Marshal(p)
	if err != nil {
		return nil, false
	}
	return json.RawMessage(data), true
}

// SetPlayer stores a player's state.
func (gs *GameState) SetPlayer(playerID string, state *engine.PlayerState) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.players[playerID] = state
}

// RemovePlayer removes a player's state.
func (gs *GameState) RemovePlayer(playerID string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	delete(gs.players, playerID)
}

// GetPlanet returns a specific planet's state (read-only copy).
func (gs *GameState) GetPlanet(playerID, planetID string) (*engine.PlanetState, bool) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	p, ok := gs.players[playerID]
	if !ok {
		return nil, false
	}
	planet, ok := p.Planets[planetID]
	if !ok {
		return nil, false
	}
	data, _ := json.Marshal(planet)
	var copy engine.PlanetState
	json.Unmarshal(data, &copy)
	return &copy, true
}

// UpdatePlanet atomically reads, modifies, and writes back a planet's state.
// The callback receives the live planet pointer — modifications persist.
// Returns false if player or planet not found.
func (gs *GameState) UpdatePlanet(playerID, planetID string, fn func(planet *engine.PlanetState) error) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	p, ok := gs.players[playerID]
	if !ok {
		return fmt.Errorf("player not found: %s", playerID)
	}
	planet, ok := p.Planets[planetID]
	if !ok {
		return fmt.Errorf("planet not found: %s", planetID)
	}
	return fn(planet)
}

// UpdatePlayer atomically reads and modifies a player's state.
func (gs *GameState) UpdatePlayer(playerID string, fn func(player *engine.PlayerState) error) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	p, ok := gs.players[playerID]
	if !ok {
		return fmt.Errorf("player not found: %s", playerID)
	}
	return fn(p)
}

// EnsurePlayer ensures a player entry exists, creating one if needed.
func (gs *GameState) EnsurePlayer(playerID string, gameSpeed int) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	if _, ok := gs.players[playerID]; !ok {
		gs.players[playerID] = &engine.PlayerState{
			ID:        playerID,
			UserID:    playerID,
			Planets:   make(map[string]*engine.PlanetState),
			Moons:     make(map[string]string),
			GameSpeed: gameSpeed,
			DebrisField: engine.Resources{},
		}
	}
}

// Tick processes all pending actions for all players.
// Holds the write lock for the entire duration — the inner work per player
// (resource arithmetic, queue timers, fleet time comparisons) is fast, and
// battles only fire when fleets arrive (infrequent). This keeps the code
// simple and correct: no data races with concurrent API reads/writes.
func (gs *GameState) Tick() {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	now := time.Now().UnixMilli()
	for _, player := range gs.players {
		gs.tickPlayer(player, now)
	}
}

// TickPlayer processes all pending actions for a specific player.
func (gs *GameState) TickPlayer(playerID string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	player, ok := gs.players[playerID]
	if !ok {
		return fmt.Errorf("player not found: %s", playerID)
	}
	now := time.Now().UnixMilli()
	gs.tickPlayer(player, now)
	return nil
}

// tickPlayer processes all pending actions for one player.
// Called with gs.mu held (write lock).
func (gs *GameState) tickPlayer(player *engine.PlayerState, now int64) {
	// Update resources for all planets.
	for _, planet := range player.Planets {
		engine.UpdatePlanetResources(planet, now, player.GameSpeed)
		gs.processBuildingQueue(planet, now, player.GameSpeed)
		gs.processResearchQueue(planet, now, player.GameSpeed)
		gs.processShipQueue(planet, now, player.GameSpeed)
		gs.processDefenseQueue(planet, now, player.GameSpeed)
	}

	// Process fleet missions.
	for i := range player.FleetMissions {
		mission := &player.FleetMissions[i]

		if mission.Status == "outbound" && now >= mission.ArrivalTime {
			gs.processFleetArrival(player, mission, now)
		} else if mission.Status == "returning" && now >= mission.ReturnTime {
			gs.processFleetReturn(player, mission, now)
		}
	}

	// Clean up completed missions.
	active := player.FleetMissions[:0]
	for _, m := range player.FleetMissions {
		if m.Status != "completed" {
			active = append(active, m)
		}
	}
	player.FleetMissions = active
}

// processFleetArrival handles a fleet mission that has arrived at its target.
// Called with gs.mu held (write lock).
func (gs *GameState) processFleetArrival(player *engine.PlayerState, mission *engine.FleetMission, now int64) {
	switch mission.MissionType {
	case "attack":
		defenderPlanet := gs.findPlanetByCoord(mission.Target)
		if defenderPlanet == nil {
			mission.Status = "returning"
			mission.ReturnTime = mission.ArrivalTime + (mission.ArrivalTime - mission.DepartureTime)
			break
		}

		var originPlanet *engine.PlanetState
		for _, p := range player.Planets {
			if p.Coordinate == mission.Origin {
				originPlanet = p
				break
			}
		}
		if originPlanet == nil {
			mission.Status = "returning"
			mission.ReturnTime = mission.ArrivalTime + (mission.ArrivalTime - mission.DepartureTime)
			break
		}

		attackerSide := engine.BattleSide{
			Ships:      mission.Fleet,
			Defense:    map[string]int{},
			WeaponTech: originPlanet.Technologies["weaponsTech"],
			ShieldTech: originPlanet.Technologies["shieldingTech"],
			ArmorTech:  originPlanet.Technologies["armorTech"],
		}
		defenderSide := engine.BattleSide{
			Ships:             defenderPlanet.Ships,
			Defense:           defenderPlanet.Defenses,
			WeaponTech:        defenderPlanet.Technologies["weaponsTech"],
			ShieldTech:        defenderPlanet.Technologies["shieldingTech"],
			ArmorTech:         defenderPlanet.Technologies["armorTech"],
			DefenderResources: defenderPlanet.Resources,
		}

		result := engine.SimulateBattle(attackerSide, defenderSide, 0)

		mission.Fleet = result.AttackerRemaining
		defenderPlanet.Ships = result.DefenderFleetRemaining
		defenderPlanet.Defenses = result.DefenderDefenseRemaining
		mission.Cargo = mission.Cargo.Add(result.Plunder)
		defenderPlanet.Resources = defenderPlanet.Resources.Sub(result.Plunder)
		player.DebrisField = player.DebrisField.Add(result.DebrisField)

		gs.emitEvent(GameEvent{
			Type:     "battleResult",
			PlayerID: player.ID,
			Data: map[string]interface{}{
				"missionId": mission.ID,
				"target":    mission.Target,
				"winner":    result.Winner,
				"rounds":    result.Rounds,
				"plunder":   result.Plunder,
				"debris":    result.DebrisField,
				"remaining": result.AttackerRemaining,
			},
		})

		mission.Status = "returning"
		mission.ReturnTime = mission.ArrivalTime + (mission.ArrivalTime - mission.DepartureTime)

	case "deploy":
		targetPlanet := gs.findPlanetByCoord(mission.Target)
		if targetPlanet != nil {
			for shipType, count := range mission.Fleet {
				targetPlanet.Ships[shipType] += count
			}
			targetPlanet.Resources = targetPlanet.Resources.Add(mission.Cargo)
		}
		mission.Fleet = map[string]int{}
		mission.Cargo = engine.Resources{}
		mission.Status = "completed"

	case "transport":
		targetPlanet := gs.findPlanetByCoord(mission.Target)
		if targetPlanet != nil {
			targetPlanet.Resources = targetPlanet.Resources.Add(mission.Cargo)
		}
		mission.Cargo = engine.Resources{}
		mission.Status = "returning"
		mission.ReturnTime = mission.ArrivalTime + (mission.ArrivalTime - mission.DepartureTime)

	case "spy":
		mission.Status = "returning"
		mission.ReturnTime = mission.ArrivalTime + (mission.ArrivalTime - mission.DepartureTime)

	case "recycle":
		mission.Status = "returning"
		mission.ReturnTime = mission.ArrivalTime + (mission.ArrivalTime - mission.DepartureTime)

	case "colonize":
		mission.Status = "returning"
		mission.ReturnTime = mission.ArrivalTime + (mission.ArrivalTime - mission.DepartureTime)

	case "expedition":
		mission.Status = "returning"
		mission.ReturnTime = mission.ArrivalTime + (mission.ArrivalTime - mission.DepartureTime)

	default:
		mission.Status = "returning"
		mission.ReturnTime = mission.ArrivalTime + (mission.ArrivalTime - mission.DepartureTime)
	}
}

func (gs *GameState) processBuildingQueue(planet *engine.PlanetState, now int64, gameSpeed int) {
	if len(planet.BuildingQueue) == 0 {
		return
	}

	item := &planet.BuildingQueue[0]
	if now >= item.EndTime {
		// Building complete
		currentLevel := planet.Buildings[item.Type]
		planet.Buildings[item.Type] = currentLevel + 1

		// Remove completed item
		planet.BuildingQueue = planet.BuildingQueue[1:]

		// Recalculate storage capacity
		storageLevel := planet.Buildings["metalStorage"]
		planet.StorageCap.Metal = engine.CalculateResourceCapacity(storageLevel, 1.0)
		storageLevel = planet.Buildings["crystalStorage"]
		planet.StorageCap.Crystal = engine.CalculateResourceCapacity(storageLevel, 1.0)
		storageLevel = planet.Buildings["deuteriumTank"]
		planet.StorageCap.Deuterium = engine.CalculateResourceCapacity(storageLevel, 1.0)
		storageLevel = planet.Buildings["darkMatterTank"]
		planet.StorageCap.DarkMatter = engine.CalculateResourceCapacity(storageLevel, 1.0)

		gs.emitEvent(GameEvent{
			Type: "buildingComplete",
			Data: map[string]interface{}{"planetId": planet.ID, "building": item.Type, "level": currentLevel + 1},
		})
	}
}

func (gs *GameState) processResearchQueue(planet *engine.PlanetState, now int64, gameSpeed int) {
	if len(planet.ResearchQueue) == 0 {
		return
	}

	item := &planet.ResearchQueue[0]
	if now >= item.EndTime {
		// Research complete
		currentLevel := planet.Technologies[item.Type]
		planet.Technologies[item.Type] = currentLevel + 1

		// Remove completed item
		planet.ResearchQueue = planet.ResearchQueue[1:]

		gs.emitEvent(GameEvent{
			Type: "researchComplete",
			Data: map[string]interface{}{"planetId": planet.ID, "type": item.Type, "level": currentLevel + 1},
		})
	}
}

func (gs *GameState) processShipQueue(planet *engine.PlanetState, now int64, gameSpeed int) {
	if len(planet.ShipQueue) == 0 {
		return
	}

	item := &planet.ShipQueue[0]
	if now >= item.EndTime {
		// Ship production complete
		currentCount := planet.Ships[item.Type]
		planet.Ships[item.Type] = currentCount + item.Count

		// Remove completed item
		planet.ShipQueue = planet.ShipQueue[1:]

		gs.emitEvent(GameEvent{
			Type: "shipComplete",
			Data: map[string]interface{}{"planetId": planet.ID, "type": item.Type, "count": item.Count},
		})
	}
}

func (gs *GameState) processDefenseQueue(planet *engine.PlanetState, now int64, gameSpeed int) {
	if len(planet.DefenseQueue) == 0 {
		return
	}

	item := &planet.DefenseQueue[0]
	if now >= item.EndTime {
		// Defense production complete
		currentCount := planet.Defenses[item.Type]
		planet.Defenses[item.Type] = currentCount + item.Count

		// Remove completed item
		planet.DefenseQueue = planet.DefenseQueue[1:]

		gs.emitEvent(GameEvent{
			Type: "defenseComplete",
			Data: map[string]interface{}{"planetId": planet.ID, "type": item.Type, "count": item.Count},
		})
	}
}

// findPlanetByCoord searches all players' planets for a matching coordinate.
func (gs *GameState) findPlanetByCoord(coord engine.Coordinate) *engine.PlanetState {
	for _, p := range gs.players {
		for _, planet := range p.Planets {
			if planet.Coordinate == coord {
				return planet
			}
		}
	}
	return nil
}

func (gs *GameState) processFleetReturn(player *engine.PlayerState, mission *engine.FleetMission, now int64) {
	// Return fleet to origin planet
	originPlanet, ok := player.Planets[fmt.Sprintf("%d-%d-%d", mission.Origin.Galaxy, mission.Origin.System, mission.Origin.Position)]
	if !ok {
		// Find by coordinate
		for _, p := range player.Planets {
			if p.Coordinate == mission.Origin {
				originPlanet = p
				ok = true
				break
			}
		}
	}

	if ok {
		// Add ships back
		for shipType, count := range mission.Fleet {
			originPlanet.Ships[shipType] += count
		}
		// Add cargo back
		originPlanet.Resources = originPlanet.Resources.Add(mission.Cargo)
	}

	mission.Status = "completed"
}
