package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

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

		// Migration: add guest support columns to users (safe to run on existing DBs)
		`ALTER TABLE users ADD COLUMN is_guest BOOLEAN DEFAULT 0`,
		`ALTER TABLE users ADD COLUMN device_id TEXT DEFAULT ''`,
		`CREATE INDEX IF NOT EXISTS idx_users_device_guest ON users(device_id, is_guest)`,

		// Alliance tables
		`CREATE TABLE IF NOT EXISTS alliances (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			tag TEXT UNIQUE NOT NULL,
			description TEXT DEFAULT '',
			leader_id TEXT NOT NULL,
			max_members INTEGER DEFAULT 30,
			auto_accept BOOLEAN DEFAULT 0,
			require_approval BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS alliance_members (
			alliance_id TEXT NOT NULL,
			player_id TEXT PRIMARY KEY,
			role TEXT NOT NULL DEFAULT 'member',
			joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (alliance_id) REFERENCES alliances(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS alliance_requests (
			id TEXT PRIMARY KEY,
			alliance_id TEXT NOT NULL,
			player_id TEXT NOT NULL,
			message TEXT DEFAULT '',
			status TEXT NOT NULL DEFAULT 'pending',
			type TEXT NOT NULL DEFAULT 'join',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (alliance_id) REFERENCES alliances(id) ON DELETE CASCADE
		)`,

		// Chat messages
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id TEXT PRIMARY KEY,
			channel TEXT NOT NULL,
			sender_id TEXT NOT NULL,
			sender_name TEXT NOT NULL,
			sender_tag TEXT DEFAULT '',
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Alliance + Chat indexes
		`CREATE INDEX IF NOT EXISTS idx_alliance_members_alliance ON alliance_members(alliance_id)`,
		`CREATE INDEX IF NOT EXISTS idx_alliance_requests_alliance ON alliance_requests(alliance_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_alliance_requests_player ON alliance_requests(player_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_messages_channel ON chat_messages(channel, created_at DESC)`,

		// Leaderboard alliance columns (added for alliance tag display)
		`ALTER TABLE leaderboard ADD COLUMN alliance_tag TEXT DEFAULT ''`,
		`ALTER TABLE leaderboard ADD COLUMN alliance_name TEXT DEFAULT ''`,

		// Display name (nickname) support
		`ALTER TABLE users ADD COLUMN display_name TEXT DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN rename_count INTEGER DEFAULT 0`,
	}

	for _, m := range migrations {
		if _, err := DB.Exec(m); err != nil {
			// ALTER TABLE ADD COLUMN is not idempotent in SQLite;
			// skip "duplicate column name" errors so restarts are safe.
			if !strings.Contains(err.Error(), "duplicate column name") {
				return fmt.Errorf("migration failed: %w\nSQL: %s", err, m)
			}
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
