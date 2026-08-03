package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	LastLogin    time.Time `json:"last_login"`
	IsActive     bool      `json:"is_active"`
}

type PlayerSave struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	SaveSlot     string    `json:"save_slot"`
	GameData     string    `json:"game_data"`
	NPCData      string    `json:"npc_data"`
	UniverseData string    `json:"universe_data"`
	SavedAt      time.Time `json:"saved_at"`
}

type LeaderboardEntry struct {
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	TotalPoints   int64     `json:"total_points"`
	FleetPoints   int64     `json:"fleet_points"`
	ResearchPoints int64    `json:"research_points"`
	UpdatedAt     time.Time `json:"updated_at"`
}
