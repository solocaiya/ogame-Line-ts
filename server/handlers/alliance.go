package handlers

import (
	"database/sql"
	"net/http"
	"regexp"
	"strings"
	"time"

	"ogame-server/models"
	"ogame-server/ws"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var tagRegex = regexp.MustCompile(`^[A-Za-z0-9]{3,8}$`)

// AllianceHandler handles alliance CRUD and member management.
type AllianceHandler struct {
	db    *sql.DB
	wsHub *ws.Hub
}

// NewAllianceHandler creates a new AllianceHandler.
func NewAllianceHandler(db *sql.DB, wsHub *ws.Hub) *AllianceHandler {
	return &AllianceHandler{db: db, wsHub: wsHub}
}

// Create creates a new alliance. The creator becomes the Leader.
func (h *AllianceHandler) Create(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		Name        string `json:"name" binding:"required,min=3,max=32"`
		Tag         string `json:"tag" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Tag = strings.ToUpper(strings.TrimSpace(req.Tag))
	if !tagRegex.MatchString(req.Tag) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag must be 3-8 alphanumeric characters"})
		return
	}

	// Check if player is already in an alliance
	var existingAllianceID string
	err := h.db.QueryRow("SELECT alliance_id FROM alliance_members WHERE player_id = ?", playerID).Scan(&existingAllianceID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already in an alliance"})
		return
	}

	// Get player username
	var username string
	h.db.QueryRow("SELECT username FROM users WHERE id = ?", playerID).Scan(&username)

	allianceID := uuid.New().String()
	now := time.Now()

	alliance := models.Alliance{
		ID:              allianceID,
		Name:            req.Name,
		Tag:             req.Tag,
		Description:     req.Description,
		LeaderID:        playerID,
		MaxMembers:      30,
		AutoAccept:      false,
		RequireApproval: true,
		CreatedAt:       now,
	}

	_, err = h.db.Exec(
		`INSERT INTO alliances (id, name, tag, description, leader_id, max_members, auto_accept, require_approval, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		alliance.ID, alliance.Name, alliance.Tag, alliance.Description,
		alliance.LeaderID, alliance.MaxMembers, alliance.AutoAccept, alliance.RequireApproval, alliance.CreatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "alliance name or tag already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create alliance"})
		return
	}

	// Add creator as leader member
	_, err = h.db.Exec(
		`INSERT INTO alliance_members (alliance_id, player_id, role, joined_at) VALUES (?, ?, ?, ?)`,
		allianceID, playerID, "leader", now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add leader as member"})
		return
	}

	member := models.AllianceMember{
		AllianceID: allianceID,
		PlayerID:   playerID,
		Role:       "leader",
		JoinedAt:   now,
	}

	c.JSON(http.StatusCreated, gin.H{
		"alliance": gin.H{
			"id":               alliance.ID,
			"name":             alliance.Name,
			"tag":              alliance.Tag,
			"description":      alliance.Description,
			"leader_id":        alliance.LeaderID,
			"max_members":      alliance.MaxMembers,
			"auto_accept":      alliance.AutoAccept,
			"require_approval": alliance.RequireApproval,
			"created_at":       alliance.CreatedAt.UnixMilli(),
			"members": []gin.H{
				{
					"playerId":  member.PlayerID,
					"username":  username,
					"role":      member.Role,
					"joinedAt":  member.JoinedAt.UnixMilli(),
				},
			},
			"pendingRequests": []interface{}{},
		},
	})
}

// GetMyAlliance returns the caller's alliance with members and pending requests.
func (h *AllianceHandler) GetMyAlliance(c *gin.Context) {
	playerID := c.GetString("user_id")

	var allianceID string
	var role string
	err := h.db.QueryRow(
		"SELECT alliance_id, role FROM alliance_members WHERE player_id = ?", playerID,
	).Scan(&allianceID, &role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not in an alliance"})
		return
	}

	alliance, err := h.getAllianceFull(allianceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load alliance"})
		return
	}

	c.JSON(http.StatusOK, alliance)
}

// Search searches alliances by name or tag.
func (h *AllianceHandler) Search(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}

	rows, err := h.db.Query(
		`SELECT id, name, tag, description, leader_id, max_members, auto_accept, require_approval, created_at
		 FROM alliances WHERE name LIKE ? OR tag LIKE ? ORDER BY created_at DESC LIMIT 50`,
		"%"+query+"%", "%"+query+"%",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}
	defer rows.Close()

	var results []gin.H
	for rows.Next() {
		var a models.Alliance
		if err := rows.Scan(&a.ID, &a.Name, &a.Tag, &a.Description, &a.LeaderID, &a.MaxMembers, &a.AutoAccept, &a.RequireApproval, &a.CreatedAt); err != nil {
			continue
		}
		// Count members
		var memberCount int
		h.db.QueryRow("SELECT COUNT(*) FROM alliance_members WHERE alliance_id = ?", a.ID).Scan(&memberCount)

		results = append(results, gin.H{
			"id":               a.ID,
			"name":             a.Name,
			"tag":              a.Tag,
			"description":      a.Description,
			"leader_id":        a.LeaderID,
			"max_members":      a.MaxMembers,
			"auto_accept":      a.AutoAccept,
			"require_approval": a.RequireApproval,
			"created_at":       a.CreatedAt.UnixMilli(),
			"member_count":     memberCount,
		})
	}

	if results == nil {
		results = []gin.H{}
	}
	c.JSON(http.StatusOK, gin.H{"alliances": results})
}

// UpdateSettings updates alliance settings (Leader/Officer only).
func (h *AllianceHandler) UpdateSettings(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		Description     *string `json:"description"`
		MaxMembers      *int    `json:"maxMembers"`
		AutoAccept      *bool   `json:"autoAccept"`
		RequireApproval *bool   `json:"requireApproval"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allianceID, role, err := h.getMemberInfo(playerID)
	if err != nil || (role != "leader" && role != "officer") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	if req.Description != nil {
		h.db.Exec("UPDATE alliances SET description = ? WHERE id = ?", *req.Description, allianceID)
	}
	if req.MaxMembers != nil && *req.MaxMembers > 0 && *req.MaxMembers <= 100 {
		h.db.Exec("UPDATE alliances SET max_members = ? WHERE id = ?", *req.MaxMembers, allianceID)
	}
	if req.AutoAccept != nil {
		h.db.Exec("UPDATE alliances SET auto_accept = ? WHERE id = ?", *req.AutoAccept, allianceID)
	}
	if req.RequireApproval != nil {
		h.db.Exec("UPDATE alliances SET require_approval = ? WHERE id = ?", *req.RequireApproval, allianceID)
	}

	// Notify alliance members
	h.notifyAllianceMembers(allianceID, "alliance:settings_updated", gin.H{
		"description":      req.Description,
		"maxMembers":       req.MaxMembers,
		"autoAccept":       req.AutoAccept,
		"requireApproval":  req.RequireApproval,
	})

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// RequestJoin submits a request to join an alliance.
func (h *AllianceHandler) RequestJoin(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		AllianceID string `json:"allianceId" binding:"required"`
		Message    string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check not already in alliance
	var existing string
	err := h.db.QueryRow("SELECT alliance_id FROM alliance_members WHERE player_id = ?", playerID).Scan(&existing)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already in an alliance"})
		return
	}

	// Check alliance exists and get settings
	var autoAccept bool
	var leaderID string
	err = h.db.QueryRow(
		"SELECT auto_accept, leader_id FROM alliances WHERE id = ?", req.AllianceID,
	).Scan(&autoAccept, &leaderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alliance not found"})
		return
	}

	// Check no duplicate pending request
	var pendingCount int
	h.db.QueryRow(
		"SELECT COUNT(*) FROM alliance_requests WHERE alliance_id = ? AND player_id = ? AND status = 'pending' AND type = 'join'",
		req.AllianceID, playerID,
	).Scan(&pendingCount)
	if pendingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "already have a pending request"})
		return
	}

	var playerName string
	h.db.QueryRow("SELECT username FROM users WHERE id = ?", playerID).Scan(&playerName)

	requestID := uuid.New().String()
	now := time.Now()

	_, err = h.db.Exec(
		`INSERT INTO alliance_requests (id, alliance_id, player_id, message, status, type, created_at)
		 VALUES (?, ?, ?, ?, 'pending', 'join', ?)`,
		requestID, req.AllianceID, playerID, req.Message, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	// If auto-accept, immediately accept
	if autoAccept {
		h.acceptJoinInternal(req.AllianceID, playerID, playerName, requestID)
		c.JSON(http.StatusOK, gin.H{"success": true, "autoAccepted": true})
		return
	}

	// Notify online officers + leader
	h.notifyAllianceOfficers(req.AllianceID, "alliance:join_request", gin.H{
		"id":         requestID,
		"playerId":   playerID,
		"playerName": playerName,
		"status":     "pending",
		"createdAt":  now.UnixMilli(),
		"message":    req.Message,
	})

	c.JSON(http.StatusOK, gin.H{"success": true, "autoAccepted": false})
}

// GetRequests returns pending join requests for the caller's alliance.
func (h *AllianceHandler) GetRequests(c *gin.Context) {
	playerID := c.GetString("user_id")

	allianceID, role, err := h.getMemberInfo(playerID)
	if err != nil || (role != "leader" && role != "officer") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	rows, err := h.db.Query(
		`SELECT r.id, r.player_id, r.message, r.status, r.created_at, u.username
		 FROM alliance_requests r JOIN users u ON r.player_id = u.id
		 WHERE r.alliance_id = ? AND r.status = 'pending' AND r.type = 'join'
		 ORDER BY r.created_at DESC`,
		allianceID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load requests"})
		return
	}
	defer rows.Close()

	var requests []gin.H
	for rows.Next() {
		var id, pid, message, status, username string
		var createdAt time.Time
		if err := rows.Scan(&id, &pid, &message, &status, &createdAt, &username); err != nil {
			continue
		}
		requests = append(requests, gin.H{
			"id":         id,
			"playerId":   pid,
			"playerName": username,
			"status":     status,
			"message":    message,
			"createdAt":  createdAt.UnixMilli(),
		})
	}

	if requests == nil {
		requests = []gin.H{}
	}
	c.JSON(http.StatusOK, gin.H{"requests": requests})
}

// AcceptRequest accepts a join request.
func (h *AllianceHandler) AcceptRequest(c *gin.Context) {
	playerID := c.GetString("user_id")
	requestID := c.Param("id")

	allianceID, role, err := h.getMemberInfo(playerID)
	if err != nil || (role != "leader" && role != "officer") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	// Get request details
	var reqPlayerID, reqStatus string
	err = h.db.QueryRow(
		"SELECT player_id, status FROM alliance_requests WHERE id = ? AND alliance_id = ?",
		requestID, allianceID,
	).Scan(&reqPlayerID, &reqStatus)
	if err != nil || reqStatus != "pending" {
		c.JSON(http.StatusNotFound, gin.H{"error": "request not found"})
		return
	}

	var playerName string
	h.db.QueryRow("SELECT username FROM users WHERE id = ?", reqPlayerID).Scan(&playerName)

	h.acceptJoinInternal(allianceID, reqPlayerID, playerName, requestID)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// RejectRequest rejects a join request.
func (h *AllianceHandler) RejectRequest(c *gin.Context) {
	playerID := c.GetString("user_id")
	requestID := c.Param("id")

	allianceID, role, err := h.getMemberInfo(playerID)
	if err != nil || (role != "leader" && role != "officer") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	h.db.Exec(
		"UPDATE alliance_requests SET status = 'rejected' WHERE id = ? AND alliance_id = ? AND status = 'pending'",
		requestID, allianceID,
	)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// InvitePlayer invites a player to the alliance.
func (h *AllianceHandler) InvitePlayer(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		PlayerID string `json:"playerId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allianceID, role, err := h.getMemberInfo(playerID)
	if err != nil || (role != "leader" && role != "officer") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	// Check target not already in alliance
	var existing string
	err = h.db.QueryRow("SELECT alliance_id FROM alliance_members WHERE player_id = ?", req.PlayerID).Scan(&existing)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "player already in an alliance"})
		return
	}

	// Check no duplicate pending invite
	var pendingCount int
	h.db.QueryRow(
		"SELECT COUNT(*) FROM alliance_requests WHERE alliance_id = ? AND player_id = ? AND status = 'pending' AND type = 'invite'",
		allianceID, req.PlayerID,
	).Scan(&pendingCount)
	if pendingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "already have a pending invite"})
		return
	}

	var inviterName, allianceName string
	h.db.QueryRow("SELECT username FROM users WHERE id = ?", playerID).Scan(&inviterName)
	h.db.QueryRow("SELECT name FROM alliances WHERE id = ?", allianceID).Scan(&allianceName)

	requestID := uuid.New().String()
	now := time.Now()

	h.db.Exec(
		`INSERT INTO alliance_requests (id, alliance_id, player_id, message, status, type, created_at)
		 VALUES (?, ?, ?, '', 'pending', 'invite', ?)`,
		requestID, allianceID, req.PlayerID, now,
	)

	// Notify the invited player
	h.wsHub.SendTo(req.PlayerID, "alliance:invite_received", gin.H{
		"requestId":    requestID,
		"allianceId":   allianceID,
		"allianceName": allianceName,
		"inviterName":  inviterName,
	})

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AcceptInvite accepts an alliance invitation.
func (h *AllianceHandler) AcceptInvite(c *gin.Context) {
	playerID := c.GetString("user_id")
	requestID := c.Param("id")

	// Get invite details
	var allianceID, reqStatus string
	err := h.db.QueryRow(
		"SELECT alliance_id, status FROM alliance_requests WHERE id = ? AND player_id = ? AND type = 'invite'",
		requestID, playerID,
	).Scan(&allianceID, &reqStatus)
	if err != nil || reqStatus != "pending" {
		c.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})
		return
	}

	// Check not already in alliance
	var existing string
	err = h.db.QueryRow("SELECT alliance_id FROM alliance_members WHERE player_id = ?", playerID).Scan(&existing)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already in an alliance"})
		return
	}

	var playerName, allianceName string
	h.db.QueryRow("SELECT username FROM users WHERE id = ?", playerID).Scan(&playerName)
	h.db.QueryRow("SELECT name FROM alliances WHERE id = ?", allianceID).Scan(&allianceName)

	h.acceptJoinInternal(allianceID, playerID, playerName, requestID)

	// Notify the player they joined
	h.wsHub.SendTo(playerID, "alliance:request_accepted", gin.H{
		"allianceId":   allianceID,
		"allianceName": allianceName,
	})

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// RejectInvite rejects an alliance invitation.
func (h *AllianceHandler) RejectInvite(c *gin.Context) {
	playerID := c.GetString("user_id")
	requestID := c.Param("id")

	h.db.Exec(
		"UPDATE alliance_requests SET status = 'rejected' WHERE id = ? AND player_id = ? AND type = 'invite' AND status = 'pending'",
		requestID, playerID,
	)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UpdateMemberRole changes a member's role (Leader only).
func (h *AllianceHandler) UpdateMemberRole(c *gin.Context) {
	playerID := c.GetString("user_id")
	targetID := c.Param("playerId")

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role != "officer" && req.Role != "member" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be officer or member"})
		return
	}

	allianceID, role, err := h.getMemberInfo(playerID)
	if err != nil || role != "leader" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only leader can change roles"})
		return
	}

	// Check target is in same alliance
	var targetAlliance string
	err = h.db.QueryRow("SELECT alliance_id FROM alliance_members WHERE player_id = ?", targetID).Scan(&targetAlliance)
	if err != nil || targetAlliance != allianceID {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not in this alliance"})
		return
	}

	h.db.Exec("UPDATE alliance_members SET role = ? WHERE player_id = ?", req.Role, targetID)

	// Notify target + officers
	h.wsHub.SendTo(targetID, "alliance:role_changed", gin.H{
		"playerId": targetID,
		"newRole":  req.Role,
	})
	h.notifyAllianceOfficers(allianceID, "alliance:role_changed", gin.H{
		"playerId": targetID,
		"newRole":  req.Role,
	})

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// RemoveMember kicks a member from the alliance.
func (h *AllianceHandler) RemoveMember(c *gin.Context) {
	playerID := c.GetString("user_id")
	targetID := c.Param("playerId")

	allianceID, role, err := h.getMemberInfo(playerID)
	if err != nil || (role != "leader" && role != "officer") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	// Can't kick leader
	var targetRole string
	err = h.db.QueryRow("SELECT role FROM alliance_members WHERE player_id = ?", targetID).Scan(&targetRole)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}
	if targetRole == "leader" {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot kick the leader"})
		return
	}
	// Officers can't kick other officers
	if targetRole == "officer" && role != "leader" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only leader can kick officers"})
		return
	}

	h.removeMemberInternal(allianceID, targetID)

	var playerName string
	h.db.QueryRow("SELECT username FROM users WHERE id = ?", targetID).Scan(&playerName)

	h.notifyAllianceMembers(allianceID, "alliance:member_left", gin.H{
		"playerId":   targetID,
		"playerName": playerName,
	})

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Leave leaves the current alliance. If leader, disbands or promotes oldest officer.
func (h *AllianceHandler) Leave(c *gin.Context) {
	playerID := c.GetString("user_id")

	allianceID, role, err := h.getMemberInfo(playerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not in an alliance"})
		return
	}

	var playerName string
	h.db.QueryRow("SELECT username FROM users WHERE id = ?", playerID).Scan(&playerName)

	if role == "leader" {
		// Try to promote oldest officer
		var oldestOfficer string
		err = h.db.QueryRow(
			"SELECT player_id FROM alliance_members WHERE alliance_id = ? AND role = 'officer' ORDER BY joined_at ASC LIMIT 1",
			allianceID,
		).Scan(&oldestOfficer)
		if err == nil {
			// Promote oldest officer to leader
			h.db.Exec("UPDATE alliance_members SET role = 'leader' WHERE player_id = ?", oldestOfficer)
			h.db.Exec("UPDATE alliances SET leader_id = ? WHERE id = ?", oldestOfficer, allianceID)
			h.notifyAllianceMembers(allianceID, "alliance:role_changed", gin.H{
				"playerId": oldestOfficer,
				"newRole":  "leader",
			})
		} else {
			// No officers — disband
			h.db.Exec("DELETE FROM alliance_members WHERE alliance_id = ?", allianceID)
			h.db.Exec("DELETE FROM alliance_requests WHERE alliance_id = ?", allianceID)
			h.db.Exec("DELETE FROM alliances WHERE id = ?", allianceID)
			h.notifyAllianceMembers(allianceID, "alliance:disbanded", gin.H{
				"allianceId": allianceID,
			})
			// Notify the leaving leader too (they were already removed from members)
			h.wsHub.SendTo(playerID, "alliance:disbanded", gin.H{
				"allianceId": allianceID,
			})
			c.JSON(http.StatusOK, gin.H{"success": true, "disbanded": true})
			return
		}
	}

	h.removeMemberInternal(allianceID, playerID)

	h.notifyAllianceMembers(allianceID, "alliance:member_left", gin.H{
		"playerId":   playerID,
		"playerName": playerName,
	})

	c.JSON(http.StatusOK, gin.H{"success": true, "disbanded": false})
}

// --- Internal helpers ---

func (h *AllianceHandler) getMemberInfo(playerID string) (allianceID string, role string, err error) {
	err = h.db.QueryRow(
		"SELECT alliance_id, role FROM alliance_members WHERE player_id = ?", playerID,
	).Scan(&allianceID, &role)
	return
}

func (h *AllianceHandler) acceptJoinInternal(allianceID, playerID, playerName, requestID string) {
	now := time.Now()

	// Update request status
	h.db.Exec("UPDATE alliance_requests SET status = 'accepted' WHERE id = ?", requestID)

	// Add member
	h.db.Exec(
		"INSERT INTO alliance_members (alliance_id, player_id, role, joined_at) VALUES (?, ?, 'member', ?)",
		allianceID, playerID, now,
	)

	// Notify alliance members
	h.notifyAllianceMembers(allianceID, "alliance:member_joined", gin.H{
		"playerId":  playerID,
		"username":  playerName,
		"role":      "member",
		"joinedAt":  now.UnixMilli(),
	})

	// Notify the joining player
	h.wsHub.SendTo(playerID, "alliance:request_accepted", gin.H{
		"allianceId": allianceID,
	})
}

func (h *AllianceHandler) removeMemberInternal(allianceID, playerID string) {
	h.db.Exec("DELETE FROM alliance_members WHERE player_id = ?", playerID)
	// Clean up any pending requests
	h.db.Exec("DELETE FROM alliance_requests WHERE player_id = ? AND status = 'pending'", playerID)
}

func (h *AllianceHandler) getAllianceFull(allianceID string) (gin.H, error) {
	var a models.Alliance
	err := h.db.QueryRow(
		`SELECT id, name, tag, description, leader_id, max_members, auto_accept, require_approval, created_at
		 FROM alliances WHERE id = ?`, allianceID,
	).Scan(&a.ID, &a.Name, &a.Tag, &a.Description, &a.LeaderID, &a.MaxMembers, &a.AutoAccept, &a.RequireApproval, &a.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Get members
	rows, err := h.db.Query(
		`SELECT m.player_id, m.role, m.joined_at, u.username
		 FROM alliance_members m JOIN users u ON m.player_id = u.id
		 WHERE m.alliance_id = ? ORDER BY
		 CASE m.role WHEN 'leader' THEN 0 WHEN 'officer' THEN 1 ELSE 2 END, m.joined_at ASC`,
		allianceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []gin.H
	for rows.Next() {
		var pid, role, username string
		var joinedAt time.Time
		if err := rows.Scan(&pid, &role, &joinedAt, &username); err != nil {
			continue
		}
		members = append(members, gin.H{
			"playerId": pid,
			"username": username,
			"role":     role,
			"joinedAt": joinedAt.UnixMilli(),
		})
	}
	if members == nil {
		members = []gin.H{}
	}

	// Get pending requests
	reqRows, err := h.db.Query(
		`SELECT r.id, r.player_id, r.message, r.created_at, u.username
		 FROM alliance_requests r JOIN users u ON r.player_id = u.id
		 WHERE r.alliance_id = ? AND r.status = 'pending' AND r.type = 'join'
		 ORDER BY r.created_at DESC`,
		allianceID,
	)
	if err != nil {
		return nil, err
	}
	defer reqRows.Close()

	var pendingRequests []gin.H
	for reqRows.Next() {
		var id, pid, message, username string
		var createdAt time.Time
		if err := reqRows.Scan(&id, &pid, &message, &createdAt, &username); err != nil {
			continue
		}
		pendingRequests = append(pendingRequests, gin.H{
			"id":         id,
			"playerId":   pid,
			"playerName": username,
			"status":     "pending",
			"message":    message,
			"createdAt":  createdAt.UnixMilli(),
		})
	}
	if pendingRequests == nil {
		pendingRequests = []gin.H{}
	}

	return gin.H{
		"id":               a.ID,
		"name":             a.Name,
		"tag":              a.Tag,
		"description":      a.Description,
		"leader_id":        a.LeaderID,
		"max_members":      a.MaxMembers,
		"auto_accept":      a.AutoAccept,
		"require_approval": a.RequireApproval,
		"created_at":       a.CreatedAt.UnixMilli(),
		"members":          members,
		"pending_requests": pendingRequests,
	}, nil
}

func (h *AllianceHandler) notifyAllianceMembers(allianceID, msgType string, data interface{}) {
	rows, err := h.db.Query("SELECT player_id FROM alliance_members WHERE alliance_id = ?", allianceID)
	if err != nil {
		return
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err == nil {
			ids = append(ids, pid)
		}
	}
	if len(ids) > 0 {
		h.wsHub.SendToMultiple(ids, msgType, data)
	}
}

func (h *AllianceHandler) notifyAllianceOfficers(allianceID, msgType string, data interface{}) {
	rows, err := h.db.Query(
		"SELECT player_id FROM alliance_members WHERE alliance_id = ? AND role IN ('leader', 'officer')",
		allianceID,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err == nil {
			ids = append(ids, pid)
		}
	}
	if len(ids) > 0 {
		h.wsHub.SendToMultiple(ids, msgType, data)
	}
}
