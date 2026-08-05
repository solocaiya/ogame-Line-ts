package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"ogame-server/config"
	"ogame-server/database"
	"ogame-server/gamestate"
	"ogame-server/handlers"
	"ogame-server/ws"

	"github.com/gin-gonic/gin"
)

func main() {
	portFlag := flag.Int("port", -1, "Server port (default: from config)")
	flag.Parse()

	cfg := config.Load()

	port := cfg.Port
	if *portFlag > 0 {
		port = fmt.Sprintf("%d", *portFlag)
	}

	// Initialize database
	if err := database.Init(cfg.DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize game state
	gs := gamestate.New()
	gs.SetDB(database.DB)

	// Initialize WebSocket hub
	wsHub := ws.NewHub(cfg.AllowedOrigins)

	// Wire WS connect/disconnect to active player tracking (optimization #3)
	wsHub.OnConnect = func(playerID string) {
		gs.MarkActive(playerID)
	}
	wsHub.OnDisconnect = func(playerID string) {
		gs.MarkInactive(playerID)
	}

	go wsHub.Run()
	log.Println("WebSocket hub started")

	// Wire game events to WebSocket broadcasts
	gs.SetEventHandler(func(event gamestate.GameEvent) {
		if event.PlayerID != "" {
			wsHub.SendTo(event.PlayerID, event.Type, event.Data)
		} else {
			wsHub.Broadcast(event.Type, event.Data)
		}
	})

	// Create a context that cancels on SIGTERM/SIGINT for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start tick loop (1 tick per second)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				gs.Tick()
			}
		}
	}()
	log.Println("Game tick loop started (1s interval)")

	// Periodic save (every 30 seconds)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := gs.SaveAllPlayers(database.DB); err != nil {
					log.Printf("Failed to save game states: %v", err)
				}
			}
		}
	}()
	log.Println("Periodic game state save started (30s interval)")

	// Periodic leaderboard refresh (every 60 seconds)
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		// Run once immediately at startup
		gs.CalculateLeaderboard(database.DB)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				gs.CalculateLeaderboard(database.DB)
			}
		}
	}()
	log.Println("Periodic leaderboard refresh started (60s interval)")

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(handlers.CORS(cfg.AllowedOrigins))

	// Health check (for Docker HEALTHCHECK)
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Auth routes
	authHandler := &handlers.AuthHandler{
		JWTSecret:       cfg.JWTSecret,
		AccessTokenTTL:  cfg.AccessTokenTTL,
		RefreshTokenTTL: cfg.RefreshTokenTTL,
	}

	// Rate limit auth endpoints: 10 requests per minute per IP
	loginLimiter := handlers.NewRateLimiter(10, time.Minute)

	auth := r.Group("/api/auth")
	{
		auth.POST("/register", handlers.RateLimit(loginLimiter), authHandler.Register)
		auth.POST("/login", handlers.RateLimit(loginLimiter), authHandler.Login)
		auth.POST("/refresh", handlers.RateLimit(loginLimiter), authHandler.Refresh)
		auth.POST("/guest", handlers.RateLimit(loginLimiter), authHandler.Guest)
		auth.GET("/me", handlers.AuthRequired(cfg.JWTSecret), authHandler.Me)
		auth.POST("/bind", handlers.AuthRequired(cfg.JWTSecret), authHandler.Bind)
		auth.GET("/bound", handlers.AuthRequired(cfg.JWTSecret), authHandler.Bound)
	}

	// Player save routes
	playerHandler := &handlers.PlayerHandler{}

	player := r.Group("/api/player")
	player.Use(handlers.AuthRequired(cfg.JWTSecret))
	{
		player.PUT("/save", playerHandler.SaveGame)
		player.GET("/save", playerHandler.LoadGame)
		player.GET("/saves", playerHandler.ListSaves)
		player.DELETE("/save", playerHandler.DeleteSave)
	}

	// Game routes
	gameHandler := handlers.NewGameHandler(gs, database.DB)

	game := r.Group("/api/game")
	game.Use(handlers.AuthRequired(cfg.JWTSecret))
	{
		game.POST("/init", gameHandler.InitPlayer)
		game.GET("/state", gameHandler.GetGameState)
		game.PUT("/settings", gameHandler.UpdateSettings)
		game.GET("/settings", gameHandler.GetSettings)
		game.POST("/building/start", gameHandler.StartBuilding)
		game.POST("/building/cancel", gameHandler.CancelBuilding)
		game.POST("/ship/start", gameHandler.StartShipProduction)
		game.POST("/ship/cancel", gameHandler.CancelShipProduction)
		game.POST("/defense/start", gameHandler.StartDefenseProduction)
		game.POST("/defense/cancel", gameHandler.CancelDefenseProduction)
		game.POST("/research/start", gameHandler.StartResearch)
		game.POST("/research/cancel", gameHandler.CancelResearch)
		game.POST("/fleet/send", gameHandler.SendFleet)
		game.POST("/fleet/recall", gameHandler.RecallFleet)
		game.GET("/leaderboard", gameHandler.GetLeaderboard)
		game.GET("/notifications", gameHandler.GetNotifications)
		game.POST("/notifications/:id/read", gameHandler.MarkNotificationRead)
		game.GET("/battle-replays", gameHandler.GetBattleReplays)
	}

	// WebSocket endpoint
	wsGroup := r.Group("/api/ws")
	wsGroup.Use(handlers.AuthRequired(cfg.JWTSecret))
	{
		wsGroup.GET("", func(c *gin.Context) {
			playerID := c.GetString("user_id")
			wsHub.HandleWS(c, playerID)
		})
	}

	// Start HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", port),
		Handler: r,
	}

	go func() {
		log.Printf("OGame API Server starting on %s", srv.Addr)
		log.Printf("API endpoints available at http://localhost:%s/api/", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Block until signal received
	<-ctx.Done()
	log.Println("Shutdown signal received — draining connections...")

	// Final save before shutdown
	if err := gs.SaveAllPlayers(database.DB); err != nil {
		log.Printf("Final save failed: %v", err)
	} else {
		log.Println("Final game state save complete")
	}

	// Shutdown HTTP server (10s timeout for in-flight requests)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
