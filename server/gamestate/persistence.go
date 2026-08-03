package gamestate

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"ogame-server/engine"
)

// SavePlayer persists a player's state to the database.
// Uses MarshalPlayer to serialize directly under the read lock, avoiding
// the Marshal→Unmarshal→Marshal triple-serialization of GetPlayer.
func (gs *GameState) SavePlayer(db *sql.DB, playerID string) error {
	data, ok := gs.MarshalPlayer(playerID)
	if !ok {
		return fmt.Errorf("player not found: %s", playerID)
	}

	now := time.Now()
	_, err := db.Exec(`
		INSERT INTO player_game_states (player_id, state_data, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(player_id) DO UPDATE SET state_data = ?, updated_at = ?
	`, playerID, string(data), now, string(data), now)

	return err
}

// LoadPlayer loads a player's state from the database.
func (gs *GameState) LoadPlayer(db *sql.DB, playerID string) error {
	var data string
	err := db.QueryRow(`SELECT state_data FROM player_game_states WHERE player_id = ?`, playerID).Scan(&data)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no saved state for player: %s", playerID)
	}
	if err != nil {
		return fmt.Errorf("query player state: %w", err)
	}

	var player engine.PlayerState
	if err := json.Unmarshal([]byte(data), &player); err != nil {
		return fmt.Errorf("unmarshal player state: %w", err)
	}

	gs.SetPlayer(playerID, &player)
	return nil
}

// SaveAllPlayers persists all players' states to the database.
// Logs individual errors and continues saving remaining players, returning
// an aggregated error at the end so one failure doesn't skip the rest.
func (gs *GameState) SaveAllPlayers(db *sql.DB) error {
	gs.mu.RLock()
	playerIDs := make([]string, 0, len(gs.players))
	for id := range gs.players {
		playerIDs = append(playerIDs, id)
	}
	gs.mu.RUnlock()

	var errs []string
	for _, id := range playerIDs {
		if err := gs.SavePlayer(db, id); err != nil {
			log.Printf("Failed to save player %s: %v", id, err)
			errs = append(errs, fmt.Sprintf("%s: %v", id, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to save %d/%d players: %s", len(errs), len(playerIDs), strings.Join(errs, "; "))
	}
	return nil
}

// InitGameTable creates the game state table if it doesn't exist.
func InitGameTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS player_game_states (
			player_id TEXT PRIMARY KEY,
			state_data TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}
