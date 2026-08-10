package models

import "time"

// User represents a player account (registered or guest)
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	LastLogin    time.Time `json:"last_login"`
	IsActive     bool      `json:"is_active"`
	IsGuest      bool      `json:"is_guest"`
	DeviceID     string    `json:"device_id,omitempty"`
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

// Alliance represents a player alliance/guild
type Alliance struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Tag              string    `json:"tag"`
	Description      string    `json:"description"`
	LeaderID         string    `json:"leader_id"`
	MaxMembers       int       `json:"max_members"`
	AutoAccept       bool      `json:"auto_accept"`
	RequireApproval  bool      `json:"require_approval"`
	CreatedAt        time.Time `json:"created_at"`
}

// AllianceMember represents a player's membership in an alliance
type AllianceMember struct {
	AllianceID string    `json:"alliance_id"`
	PlayerID   string    `json:"player_id"`
	Role       string    `json:"role"` // leader, officer, member
	JoinedAt   time.Time `json:"joined_at"`
}

// AllianceRequest represents a join request or invitation
type AllianceRequest struct {
	ID         string    `json:"id"`
	AllianceID string    `json:"alliance_id"`
	PlayerID   string    `json:"player_id"`
	Message    string    `json:"message"`
	Status     string    `json:"status"` // pending, accepted, rejected
	Type       string    `json:"type"`   // join, invite
	CreatedAt  time.Time `json:"created_at"`
}

// ChatMessage represents a chat message in world or alliance channel
type ChatMessage struct {
	ID         string    `json:"id"`
	Channel    string    `json:"channel"` // world, alliance
	SenderID   string    `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	SenderTag  string    `json:"sender_tag,omitempty"` // alliance tag
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}
