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
	NATSURL                      string
	LogLevel                     string
	AutoMigrate                  bool
	ShutdownTimeout              time.Duration
	ClawmemBaseURL               string
	ClawmemTimeout               time.Duration
	IdentityBaseURL              string
	IdentityTenant               string
	IdentityTimeout              time.Duration
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
	DefaultExecutionRing         string
	PolicyBundleID               string
	PolicyBundleVersion          string
}

func LoadServerFromEnv() (Server, error) {
	cfg := Server{
		AppEnv:                       envOrDefault("APP_ENV", "development"),
		HTTPAddress:                  envOrDefault("SERVER_ADDRESS", "127.0.0.1:8080"),
		DatabaseURL:                  strings.TrimSpace(os.Getenv("DATABASE_URL")),
		NATSURL:                      envOrDefault("CLAWBOT_NATS_URL", envOrDefault("NATS_URL", "nats://127.0.0.1:4222")),
		LogLevel:                     envOrDefault("LOG_LEVEL", "info"),
		AutoMigrate:                  parseBool(envOrDefault("AUTO_MIGRATE", "true")),
		ClawmemBaseURL:               strings.TrimSpace(os.Getenv("CLAWMEM_BASE_URL")),
		InferenceBaseURL:             envOrDefault("INFERENCE_BASE_URL", "http://ai-precision:11434"),
		IdentityBaseURL:              strings.TrimSpace(os.Getenv("CLAWBOT_IDENTITY_BASE_URL")),
		IdentityTenant:               strings.TrimSpace(os.Getenv("CLAWBOT_IDENTITY_TENANT")),
		ModelProvider:                envOrDefault("MODEL_PROVIDER", "local_ollama"),
		PrimaryModel:                 envOrDefault("PRIMARY_MODEL", "ibm/granite3.3:8b"),
		GuardrailModel:               envOrDefault("GUARDRAIL_MODEL", "ibm/granite3.3-guardian:8b"),
		HelperModel:                  envOrDefault("HELPER_MODEL", "granite4:3b"),
		DisableLocalOllamaGuardrails: parseBool(envOrDefault("LOCAL_OLLAMA_DISABLE_GUARDRAILS", "false")),
		EnableCompactDualPayload:     parseBool(envOrDefault("ENABLE_COMPACT_DUAL_PAYLOAD", "true")),
		DefaultExecutionRing:         envOrDefault("DEFAULT_EXECUTION_RING", "ring_1"),
		PolicyBundleID:               envOrDefault("POLICY_BUNDLE_ID", "ach-governance"),
		PolicyBundleVersion:          envOrDefault("POLICY_BUNDLE_VERSION", "2026.1"),
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

	identityTimeout, err := time.ParseDuration(envOrDefault("CLAWBOT_IDENTITY_TIMEOUT", "5s"))
	if err != nil {
		return Server{}, fmt.Errorf("parse CLAWBOT_IDENTITY_TIMEOUT: %w", err)
	}
	cfg.IdentityTimeout = identityTimeout

	inferenceTimeout, err := time.ParseDuration(envOrDefault("INFERENCE_TIMEOUT", "120s"))
	if err != nil {
		return Server{}, fmt.Errorf("parse INFERENCE_TIMEOUT: %w", err)
	}
	cfg.InferenceTimeout = inferenceTimeout

	guardrailTimeout, err := time.ParseDuration(envOrDefault("GUARDRAIL_TIMEOUT", "30s"))
	if err != nil {
		return Server{}, fmt.Errorf("parse GUARDRAIL_TIMEOUT: %w", err)
	}
	cfg.GuardrailTimeout = guardrailTimeout

	helperTimeout, err := time.ParseDuration(envOrDefault("HELPER_TIMEOUT", "30s"))
	if err != nil {
		return Server{}, fmt.Errorf("parse HELPER_TIMEOUT: %w", err)
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
