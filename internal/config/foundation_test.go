package config

import "testing"

func TestLoadFoundationFromEnvDefaults(t *testing.T) {
	t.Setenv("STACK_SMOKE_TIMEOUT", "7s")

	cfg, err := LoadFoundationFromEnv()
	if err != nil {
		t.Fatalf("LoadFoundationFromEnv() error = %v", err)
	}

	if cfg.PostgresPort != "5432" {
		t.Fatalf("expected default postgres port, got %q", cfg.PostgresPort)
	}

	if cfg.ZeroClawURL != "http://127.0.0.1:3000/health" {
		t.Fatalf("unexpected zeroclaw url: %q", cfg.ZeroClawURL)
	}

	if cfg.Timeout.String() != "7s" {
		t.Fatalf("unexpected timeout: %s", cfg.Timeout)
	}
}

func TestLoadFoundationFromEnvInvalidTimeout(t *testing.T) {
	t.Setenv("STACK_SMOKE_TIMEOUT", "definitely-not-a-duration")

	if _, err := LoadFoundationFromEnv(); err == nil {
		t.Fatal("expected timeout parse error")
	}
}
