package config

import (
	"os"
	"testing"
)

func clearServerEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"DATABASE_URL",
		"SERVER_ADDRESS",
		"CONTROL_PLANE_ENABLED",
		"STACK_SMOKE_TIMEOUT",
		"SHUTDOWN_TIMEOUT",
		"CLAWMEM_BASE_URL",
		"CLAWMEM_TIMEOUT",
		"INFERENCE_BASE_URL",
		"INFERENCE_PROVIDER",
		"INFERENCE_MODEL_PROFILE",
		"INFERENCE_TIMEOUT",
		"GUARDRAIL_TIMEOUT",
		"HELPER_TIMEOUT",
		"GUARDRAIL_MODEL",
		"HELPER_MODEL",
		"LOCAL_OLLAMA_DISABLE_GUARDRAILS",
		"ENABLE_COMPACT_DUAL_PAYLOAD",
	}

	saved := make(map[string]*string, len(keys))
	for _, k := range keys {
		if v, ok := os.LookupEnv(k); ok {
			vv := v
			saved[k] = &vv
		} else {
			saved[k] = nil
		}
		_ = os.Unsetenv(k)
	}

	t.Cleanup(func() {
		for _, k := range keys {
			if v := saved[k]; v != nil {
				_ = os.Setenv(k, *v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	})
}

func TestLoadServerFromEnvDefaults(t *testing.T) {
	clearServerEnv(t)
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
	if cfg.InferenceTimeout.String() != "2m0s" {
		t.Fatalf("unexpected InferenceTimeout: %s", cfg.InferenceTimeout)
	}
	if cfg.GuardrailTimeout.String() != "30s" {
		t.Fatalf("unexpected GuardrailTimeout: %s", cfg.GuardrailTimeout)
	}
	if cfg.HelperTimeout.String() != "30s" {
		t.Fatalf("unexpected HelperTimeout: %s", cfg.HelperTimeout)
	}
	if !cfg.EnableCompactDualPayload {
		t.Fatal("expected EnableCompactDualPayload to default to true")
	}
	if cfg.DisableLocalOllamaGuardrails {
		t.Fatal("expected DisableLocalOllamaGuardrails to default to false")
	}
}

func TestLoadServerFromEnvRequiresDatabaseURL(t *testing.T) {
	clearServerEnv(t)

	if _, err := LoadServerFromEnv(); err == nil {
		t.Fatal("expected DATABASE_URL validation error")
	}
}

func TestLoadServerFromEnvInvalidShutdownTimeout(t *testing.T) {
	clearServerEnv(t)
	t.Setenv("DATABASE_URL", "postgres://clawbot:test@127.0.0.1:5432/clawbot?sslmode=disable")
	t.Setenv("SHUTDOWN_TIMEOUT", "later")

	if _, err := LoadServerFromEnv(); err == nil {
		t.Fatal("expected SHUTDOWN_TIMEOUT parse error")
	}
}

func TestLoadServerFromEnvInvalidOptionalDurations(t *testing.T) {
	clearServerEnv(t)
	t.Setenv("DATABASE_URL", "postgres://clawbot:test@127.0.0.1:5432/clawbot?sslmode=disable")
	t.Setenv("GUARDRAIL_TIMEOUT", "not-a-duration")

	if _, err := LoadServerFromEnv(); err == nil {
		t.Fatal("expected GUARDRAIL_TIMEOUT parse error")
	}
}
