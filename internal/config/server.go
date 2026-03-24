package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Server struct {
	AppEnv          string
	HTTPAddress     string
	DatabaseURL     string
	LogLevel        string
	AutoMigrate     bool
	ShutdownTimeout time.Duration
}

func LoadServerFromEnv() (Server, error) {
	cfg := Server{
		AppEnv:      envOrDefault("APP_ENV", "development"),
		HTTPAddress: envOrDefault("SERVER_ADDRESS", "127.0.0.1:8080"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		LogLevel:    envOrDefault("LOG_LEVEL", "info"),
		AutoMigrate: parseBool(envOrDefault("AUTO_MIGRATE", "true")),
	}

	shutdownTimeout, err := time.ParseDuration(envOrDefault("SHUTDOWN_TIMEOUT", "10s"))
	if err != nil {
		return Server{}, fmt.Errorf("parse SHUTDOWN_TIMEOUT: %w", err)
	}
	cfg.ShutdownTimeout = shutdownTimeout

	if cfg.DatabaseURL == "" {
		return Server{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.HTTPAddress == "" {
		return Server{}, fmt.Errorf("SERVER_ADDRESS is required")
	}

	return cfg, nil
}

func envOrDefault(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
