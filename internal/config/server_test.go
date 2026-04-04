package config

import "testing"

func TestLoadServerFromEnvDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://clawbot:test@127.0.0.1:5432/clawbot?sslmode=disable")

	cfg, err := LoadServerFromEnv()
	if err != nil {
		t.Fatalf("LoadServerFromEnv() error = %v", err)
	}

	if cfg.HTTPAddress != "127.0.0.1:8080" {
		t.Fatalf("unexpected HTTPAddress: %s", cfg.HTTPAddress)
	}

	if !cfg.AutoMigrate {
		t.Fatal("expected AutoMigrate to default to true")
	}

	if cfg.ShutdownTimeout.String() != "10s" {
		t.Fatalf("unexpected ShutdownTimeout: %s", cfg.ShutdownTimeout)
	}

	if cfg.InferenceBaseURL != "http://ai-precision:11434" {
		t.Fatalf("unexpected InferenceBaseURL: %s", cfg.InferenceBaseURL)
	}
	if !cfg.EnableCompactDualPayload {
		t.Fatal("expected EnableCompactDualPayload to default to true")
	}
	if cfg.DisableLocalOllamaGuardrails {
		t.Fatal("expected DisableLocalOllamaGuardrails to default to false")
	}
}

func TestLoadServerFromEnvRequiresDatabaseURL(t *testing.T) {
	if _, err := LoadServerFromEnv(); err == nil {
		t.Fatal("expected DATABASE_URL validation error")
	}
}

func TestLoadServerFromEnvInvalidShutdownTimeout(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://clawbot:test@127.0.0.1:5432/clawbot?sslmode=disable")
	t.Setenv("SHUTDOWN_TIMEOUT", "later")

	if _, err := LoadServerFromEnv(); err == nil {
		t.Fatal("expected SHUTDOWN_TIMEOUT parse error")
	}
}

func TestLoadServerFromEnvInvalidOptionalDurations(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://clawbot:test@127.0.0.1:5432/clawbot?sslmode=disable")
	t.Setenv("GUARDRAIL_TIMEOUT", "not-a-duration")

	if _, err := LoadServerFromEnv(); err == nil {
		t.Fatal("expected GUARDRAIL_TIMEOUT parse error")
	}
}
