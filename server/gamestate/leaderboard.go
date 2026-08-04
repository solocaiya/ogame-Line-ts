package gamestate

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"ogame-server/engine"
)

// LeaderboardEntry represents a player's leaderboard standing.
type LeaderboardEntry struct {
	UserID         string `json:"userId"`
	Username       string `json:"username"`
	TotalPoints    int64  `json:"totalPoints"`
	EconomyPoints  int64  `json:"economyPoints"`
	MilitaryPoints int64  `json:"militaryPoints"`
	ResearchPoints int64  `json:"researchPoints"`
	UpdatedAt      int64  `json:"updatedAt"`
}

// CalculateLeaderboard computes points for all players and writes to DB.
// Called periodically from a goroutine.
func (gs *GameState) CalculateLeaderboard(db *sql.DB) {
	gs.mu.RLock()
	// Snapshot all players under read lock
	snapshot := make(map[string]*engine.PlayerState, len(gs.players))
	for id, p := range gs.players {
		data, err := json.Marshal(p)
		if err != nil {
			continue
		}
		var cp engine.PlayerState
		if err := json.Unmarshal(data, &cp); err != nil {
			continue
		}
		snapshot[id] = &cp
	}
	gs.mu.RUnlock()

	now := time.Now()
	var errs []string

	for playerID, player := range snapshot {
		economy, military, research := calculatePlayerPoints(player)
		total := economy + military + research

		// Look up username from users table via player_id == user_id convention
		username := playerID
		var uname string
		if err := db.QueryRow(`SELECT username FROM users WHERE id = ?`, playerID).Scan(&uname); err == nil {
			username = uname
		}

		_, err := db.Exec(`
			INSERT INTO leaderboard (user_id, username, total_points, economy_points, military_points, research_points, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(user_id) DO UPDATE SET
				username = excluded.username,
				total_points = excluded.total_points,
				economy_points = excluded.economy_points,
				military_points = excluded.military_points,
				research_points = excluded.research_points,
				updated_at = excluded.updated_at
		`, playerID, username, total, economy, military, research, now)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", playerID, err))
		}
	}

	if len(errs) > 0 {
		log.Printf("Leaderboard calculation: %d errors: %v", len(errs), errs[:min(3, len(errs))])
	}
}

// calculatePlayerPoints computes economy, military, and research points for a player.
func calculatePlayerPoints(player *engine.PlayerState) (economy, military, research int64) {
	// Economy: buildings — sum of level * 100 per building per planet
	for _, planet := range player.Planets {
		for buildingType, level := range planet.Buildings {
			economy += int64(level) * buildingBasePoints(buildingType)
		}
		// Economy also includes resources (rough proxy for economic output)
		economy += (planet.Resources.Metal + planet.Resources.Crystal + planet.Resources.Deuterium) / 1000
	}

	// Military: ships + defense
	for _, planet := range player.Planets {
		for shipType, count := range planet.Ships {
			military += int64(count) * shipPoints(shipType)
		}
		for defType, count := range planet.Defenses {
			military += int64(count) * defensePoints(defType)
		}
	}

	// Research: tech levels
	for _, planet := range player.Planets {
		for techType, level := range planet.Technologies {
			research += int64(level) * researchBasePoints(techType)
		}
		break // research is account-wide, only count once
	}

	return
}

// buildingBasePoints returns point value per building level.
func buildingBasePoints(buildingType string) int64 {
	def, ok := engine.BuildingDefs[buildingType]
	if !ok {
		return 100
	}
	// Points = base cost in resources / 100, minimum 10
	cost := def.BaseCost
	total := cost.Metal + cost.Crystal + cost.Deuterium
	points := total / 100
	if points < 10 {
		points = 10
	}
	return points
}

// shipPoints returns point value per ship unit.
func shipPoints(shipType string) int64 {
	def, ok := engine.ShipDefs[shipType]
	if !ok {
		return 10
	}
	// Points = (attack + shield + armor) / 10, minimum 1
	points := int64(def.Attack+def.Shield+def.Armor) / 10
	if points < 1 {
		points = 1
	}
	return points
}

// defensePoints returns point value per defense unit.
func defensePoints(defType string) int64 {
	def, ok := engine.DefenseDefs[defType]
	if !ok {
		return 10
	}
	points := int64(def.Attack+def.Shield+def.Armor) / 10
	if points < 1 {
		points = 1
	}
	return points
}

// researchBasePoints returns point value per research level.
func researchBasePoints(techType string) int64 {
	def, ok := engine.ResearchDefs[techType]
	if !ok {
		return 100
	}
	cost := def.BaseCost
	total := cost.Metal + cost.Crystal + cost.Deuterium
	points := total / 50
	if points < 20 {
		points = 20
	}
	return points
}

// GetLeaderboard fetches the top leaderboard entries from DB.
func GetLeaderboard(db *sql.DB, limit, offset int) ([]LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := db.Query(`
		SELECT user_id, username, total_points, economy_points, military_points, research_points, updated_at
		FROM leaderboard
		ORDER BY total_points DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		var updatedAt string
		if err := rows.Scan(&e.UserID, &e.Username, &e.TotalPoints, &e.EconomyPoints, &e.MilitaryPoints, &e.ResearchPoints, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan leaderboard row: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
