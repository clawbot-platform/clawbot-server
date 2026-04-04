package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Server struct {
	AppEnv                       string
	HTTPAddress                  string
	DatabaseURL                  string
	LogLevel                     string
	AutoMigrate                  bool
	ShutdownTimeout              time.Duration
	ClawmemBaseURL               string
	ClawmemTimeout               time.Duration
	InferenceBaseURL             string
	InferenceTimeout             time.Duration
	GuardrailTimeout             time.Duration
	HelperTimeout                time.Duration
	ModelProvider                string
	PrimaryModel                 string
	GuardrailModel               string
	HelperModel                  string
	DisableLocalOllamaGuardrails bool
	EnableCompactDualPayload     bool
}

func LoadServerFromEnv() (Server, error) {
	cfg := Server{
		AppEnv:                       envOrDefault("APP_ENV", "development"),
		HTTPAddress:                  envOrDefault("SERVER_ADDRESS", "127.0.0.1:8080"),
		DatabaseURL:                  strings.TrimSpace(os.Getenv("DATABASE_URL")),
		LogLevel:                     envOrDefault("LOG_LEVEL", "info"),
		AutoMigrate:                  parseBool(envOrDefault("AUTO_MIGRATE", "true")),
		ClawmemBaseURL:               strings.TrimSpace(os.Getenv("CLAWMEM_BASE_URL")),
		InferenceBaseURL:             envOrDefault("INFERENCE_BASE_URL", "http://ai-precision:11434"),
		ModelProvider:                envOrDefault("MODEL_PROVIDER", "local_ollama"),
		PrimaryModel:                 envOrDefault("PRIMARY_MODEL", "ibm/granite3.3:8b"),
		GuardrailModel:               envOrDefault("GUARDRAIL_MODEL", "ibm/granite3.3-guardian:8b"),
		HelperModel:                  envOrDefault("HELPER_MODEL", "granite4:3b"),
		DisableLocalOllamaGuardrails: parseBool(envOrDefault("LOCAL_OLLAMA_DISABLE_GUARDRAILS", "false")),
		EnableCompactDualPayload:     parseBool(envOrDefault("ENABLE_COMPACT_DUAL_PAYLOAD", "true")),
	}

	shutdownTimeout, err := time.ParseDuration(envOrDefault("SHUTDOWN_TIMEOUT", "10s"))
	if err != nil {
		return Server{}, fmt.Errorf("parse SHUTDOWN_TIMEOUT: %w", err)
	}
	cfg.ShutdownTimeout = shutdownTimeout

	clawmemTimeout, err := time.ParseDuration(envOrDefault("CLAWMEM_TIMEOUT", "5s"))
	if err != nil {
		return Server{}, fmt.Errorf("parse CLAWMEM_TIMEOUT: %w", err)
	}
	cfg.ClawmemTimeout = clawmemTimeout

	inferenceTimeout, err := time.ParseDuration(envOrDefault("INFERENCE_TIMEOUT", "45s"))
	if err != nil {
		return Server{}, fmt.Errorf("parse INFERENCE_TIMEOUT: %w", err)
	}
	cfg.InferenceTimeout = inferenceTimeout

	guardrailTimeout, err := parseOptionalDuration("GUARDRAIL_TIMEOUT")
	if err != nil {
		return Server{}, err
	}
	cfg.GuardrailTimeout = guardrailTimeout

	helperTimeout, err := parseOptionalDuration("HELPER_TIMEOUT")
	if err != nil {
		return Server{}, err
	}
	cfg.HelperTimeout = helperTimeout

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

func parseOptionalDuration(envKey string) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(envKey))
	if raw == "" {
		return 0, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", envKey, err)
	}
	return value, nil
}
