package gamestate

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"ogame-server/engine"
)

// SavePlayer persists a player's state to the database.
func (gs *GameState) SavePlayer(db *sql.DB, playerID string) error {
	player, ok := gs.GetPlayer(playerID)
	if !ok {
		return fmt.Errorf("player not found: %s", playerID)
	}

	data, err := json.Marshal(player)
	if err != nil {
		return fmt.Errorf("marshal player state: %w", err)
	}

	now := time.Now()
	_, err = db.Exec(`
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
func (gs *GameState) SaveAllPlayers(db *sql.DB) error {
	gs.mu.RLock()
	playerIDs := make([]string, 0, len(gs.players))
	for id := range gs.players {
		playerIDs = append(playerIDs, id)
	}
	gs.mu.RUnlock()

	for _, id := range playerIDs {
		if err := gs.SavePlayer(db, id); err != nil {
			return fmt.Errorf("save player %s: %w", id, err)
		}
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
