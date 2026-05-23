package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	Port        string
}

func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL: getEnv("DATABASE_URL", ""),
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		Port:        getEnv("PORT", "8080"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
