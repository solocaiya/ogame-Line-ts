package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"ogame-server/database"
	"ogame-server/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type GuestRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}

type BindRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         models.User  `json:"user"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := models.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
		LastLogin:    time.Now(),
		IsActive:     true,
		IsGuest:      false,
	}

	_, err = database.DB.Exec(
		`INSERT INTO users (id, username, password_hash, created_at, last_login, is_active, is_guest)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Username, user.PasswordHash, user.CreatedAt, user.LastLogin, user.IsActive, user.IsGuest,
	)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	tokens, err := h.generateTokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}
	c.JSON(http.StatusCreated, AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	var user models.User
	err := database.DB.QueryRow(
		`SELECT id, username, password_hash, created_at, last_login, is_active, is_guest, COALESCE(device_id, '')
		 FROM users WHERE username = ?`, req.Username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.LastLogin, &user.IsActive, &user.IsGuest, &user.DeviceID)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
		return
	}

	// Guest accounts cannot login via password — they must use the /auth/guest endpoint
	if user.IsGuest {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// Update last login
	database.DB.Exec(`UPDATE users SET last_login = ? WHERE id = ?`, time.Now(), user.ID)

	tokens, err := h.generateTokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}
	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         user,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// Verify token type is "refresh" to prevent access tokens from being used here
	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
		return
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	// Verify user still exists and is active
	var isActive bool
	err = database.DB.QueryRow(`SELECT is_active FROM users WHERE id = ?`, userID).Scan(&isActive)
	if err != nil || !isActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found or disabled"})
		return
	}

	tokens, err := h.generateTokens(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")

	var user models.User
	err := database.DB.QueryRow(
		`SELECT id, username, created_at, last_login, is_active, is_guest, COALESCE(device_id, '')
		 FROM users WHERE id = ?`, userID,
	).Scan(&user.ID, &user.Username, &user.CreatedAt, &user.LastLogin, &user.IsActive, &user.IsGuest, &user.DeviceID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Guest creates a guest account bound to a device ID.
// If a guest already exists for this device, it returns tokens for the existing account.
func (h *AuthHandler) Guest(c *gin.Context) {
	var req GuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Check if guest already exists for this device
	var user models.User
	err := database.DB.QueryRow(
		`SELECT id, username, created_at, last_login, is_active, is_guest, COALESCE(device_id, '')
		 FROM users WHERE device_id = ? AND is_guest = 1`, req.DeviceID,
	).Scan(&user.ID, &user.Username, &user.CreatedAt, &user.LastLogin, &user.IsActive, &user.IsGuest, &user.DeviceID)

	if err == nil {
		// Existing guest — update last login and return tokens
		database.DB.Exec(`UPDATE users SET last_login = ? WHERE id = ?`, time.Now(), user.ID)
		tokens, err := h.generateTokens(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
			return
		}
		c.JSON(http.StatusOK, AuthResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			User:         user,
		})
		return
	}

	if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Create new guest account
	user = models.User{
		ID:       uuid.New().String(),
		Username: "guest_" + uuid.New().String()[:8],
		CreatedAt: time.Now(),
		LastLogin: time.Now(),
		IsActive: true,
		IsGuest:  true,
		DeviceID: req.DeviceID,
	}

	_, err = database.DB.Exec(
		`INSERT INTO users (id, username, password_hash, created_at, last_login, is_active, is_guest, device_id)
		 VALUES (?, ?, '', ?, ?, ?, ?, ?)`,
		user.ID, user.Username, user.CreatedAt, user.LastLogin, user.IsActive, user.IsGuest, user.DeviceID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create guest account"})
		return
	}

	tokens, err := h.generateTokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}
	c.JSON(http.StatusCreated, AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         user,
	})
}

// Bind upgrades a guest account to a registered account by setting username + password.
// Returns 409 if the account is already bound (not a guest).
func (h *AuthHandler) Bind(c *gin.Context) {
	userID := c.GetString("user_id")

	var req BindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Use a transaction to prevent race conditions between username check and update
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer tx.Rollback()

	// Check if user is a guest (lock the row to prevent concurrent modifications)
	var isGuest bool
	var deviceID string
	err = tx.QueryRow(
		`SELECT is_guest, COALESCE(device_id, '') FROM users WHERE id = ?`, userID,
	).Scan(&isGuest, &deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !isGuest {
		c.JSON(http.StatusConflict, gin.H{"error": "account already bound", "already_bound": true})
		return
	}

	// Check if username is taken (within transaction for consistency)
	var exists int
	err = tx.QueryRow(`SELECT COUNT(*) FROM users WHERE username = ?`, req.Username).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if exists > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Update user: set real username + password, clear guest flag
	_, err = tx.Exec(
		`UPDATE users SET username = ?, password_hash = ?, is_guest = 0 WHERE id = ?`,
		req.Username, string(hash), userID,
	)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Return updated user with fresh tokens
	var user models.User
	database.DB.QueryRow(
		`SELECT id, username, created_at, last_login, is_active, is_guest, COALESCE(device_id, '')
		 FROM users WHERE id = ?`, userID,
	).Scan(&user.ID, &user.Username, &user.CreatedAt, &user.LastLogin, &user.IsActive, &user.IsGuest, &user.DeviceID)

	tokens, err := h.generateTokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}
	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         user,
	})
}

// Bound returns whether the current user is a guest and whether they have bound an account.
func (h *AuthHandler) Bound(c *gin.Context) {
	userID := c.GetString("user_id")

	var isGuest bool
	var username string
	err := database.DB.QueryRow(
		`SELECT is_guest, username FROM users WHERE id = ?`, userID,
	).Scan(&isGuest, &username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_guest":     isGuest,
		"is_bound":     !isGuest,
		"username":     username,
	})
}

type tokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (h *AuthHandler) generateTokens(userID string) (tokenPair, error) {
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"type":    "access",
		"exp":     time.Now().Add(h.AccessTokenTTL).Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(h.JWTSecret))
	if err != nil {
		return tokenPair{}, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"type":    "refresh",
		"exp":     time.Now().Add(h.RefreshTokenTTL).Unix(),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(h.JWTSecret))
	if err != nil {
		return tokenPair{}, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
