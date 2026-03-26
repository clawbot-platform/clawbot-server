package app

import (
	"bytes"
	"testing"

	"clawbot-server/internal/platform/store"
)

func TestNewLoggerHonorsLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("debug", &buf)
	logger.Debug("debug-enabled")

	if !bytes.Contains(buf.Bytes(), []byte("debug-enabled")) {
		t.Fatalf("expected debug log output, got %s", buf.String())
	}
}

func TestBuildServicesReturnsManagers(t *testing.T) {
	pg := store.NewPostgres(nil)
	services := buildServices(pg)

	if services.runs == nil || services.bots == nil || services.policies == nil || services.dashboard == nil {
		t.Fatalf("expected non-nil services %#v", services)
	}
}
