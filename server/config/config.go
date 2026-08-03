package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port            string
	DBPath          string
	JWTSecret       string
	AllowedOrigins  []string // empty = reflect request Origin (dev); set for production
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func Load() *Config {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// Generate a random secret for development; production MUST set JWT_SECRET
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Fatalf("Failed to generate JWT secret: %v", err)
		}
		jwtSecret = hex.EncodeToString(b)
		log.Println("WARNING: JWT_SECRET not set — using auto-generated secret. Tokens will not survive restarts. Set JWT_SECRET for production!")
	}

	var allowedOrigins []string
	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		for _, o := range strings.Split(origins, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				allowedOrigins = append(allowedOrigins, o)
			}
		}
	}

	return &Config{
		Port:            getEnv("PORT", "8081"),
		DBPath:          getEnv("DB_PATH", "./ogame.db"),
		JWTSecret:       jwtSecret,
		AllowedOrigins:  allowedOrigins,
		AccessTokenTTL:  24 * time.Hour,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
