package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connected successfully")
	return migrate()
}

func migrate() error {
	migrations := []string{
		// Core tables
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME,
			is_active BOOLEAN DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS player_saves (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			save_slot TEXT DEFAULT 'default',
			game_data TEXT NOT NULL,
			npc_data TEXT,
			universe_data TEXT,
			saved_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS player_game_states (
			player_id TEXT PRIMARY KEY,
			state_data TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS leaderboard (
			user_id TEXT PRIMARY KEY,
			username TEXT,
			total_points INTEGER DEFAULT 0,
			economy_points INTEGER DEFAULT 0,
			military_points INTEGER DEFAULT 0,
			research_points INTEGER DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Battle replays — stores battle results for history/replay
		`CREATE TABLE IF NOT EXISTS battle_replays (
			id TEXT PRIMARY KEY,
			attacker_id TEXT NOT NULL,
			target_coord TEXT NOT NULL,
			result_data TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Notifications — queued for offline players, delivered on login
		`CREATE TABLE IF NOT EXISTS notifications (
			id TEXT PRIMARY KEY,
			player_id TEXT NOT NULL,
			type TEXT NOT NULL,
			message TEXT NOT NULL,
			data TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			read BOOLEAN DEFAULT 0
		)`,

		// Indexes for query performance
		`CREATE INDEX IF NOT EXISTS idx_player_saves_user_id ON player_saves(user_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_player_saves_user_slot ON player_saves(user_id, save_slot)`,
		`CREATE INDEX IF NOT EXISTS idx_leaderboard_points ON leaderboard(total_points DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_player_game_states_updated ON player_game_states(updated_at)`,
		`CREATE INDEX IF NOT EXISTS idx_battle_replays_attacker ON battle_replays(attacker_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_player ON notifications(player_id, created_at DESC)`,
	}

	for _, m := range migrations {
		if _, err := DB.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w\nSQL: %s", err, m)
		}
	}

	log.Println("Database migrations completed")
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
