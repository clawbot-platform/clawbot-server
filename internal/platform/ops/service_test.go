package ops

import (
	"context"
	"testing"
	"time"

	"clawbot-server/internal/platform/store"
	"clawbot-server/internal/version"
)

func TestManagerOverviewAndServiceListing(t *testing.T) {
	current := time.Date(2026, 3, 26, 15, 0, 0, 0, time.UTC)
	manager := newManagerWithClock(version.Info{Version: "1.2.3"}, func() time.Time { return current })

	overview, err := manager.Overview(context.Background())
	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if overview.Status != StatusDegraded {
		t.Fatalf("expected degraded overview, got %s", overview.Status)
	}
	if overview.ServicesTotal != 3 || overview.ServicesHealthy != 2 || overview.ServicesDegraded != 1 {
		t.Fatalf("unexpected overview counts %#v", overview)
	}

	services, err := manager.ListServices(context.Background())
	if err != nil {
		t.Fatalf("ListServices() error = %v", err)
	}
	if len(services) != 3 {
		t.Fatalf("expected 3 services, got %d", len(services))
	}
	if services[0].Name != "clawbot-server" {
		t.Fatalf("expected sorted services, got first %s", services[0].Name)
	}
	if services[0].UptimeSeconds == 0 {
		t.Fatalf("expected computed uptime, got %#v", services[0])
	}
}

func TestManagerMaintenanceActions(t *testing.T) {
	current := time.Date(2026, 3, 26, 15, 0, 0, 0, time.UTC)
	manager := newManagerWithClock(version.Info{Version: "1.2.3"}, func() time.Time { return current })

	service, err := manager.SetMaintenance(context.Background(), serviceClawbotServer, "ops-user")
	if err != nil {
		t.Fatalf("SetMaintenance() error = %v", err)
	}
	if service.Status != StatusMaintenance || !service.MaintenanceMode {
		t.Fatalf("expected maintenance status, got %#v", service)
	}

	resumed, err := manager.ResumeService(context.Background(), serviceClawbotServer, "ops-user")
	if err != nil {
		t.Fatalf("ResumeService() error = %v", err)
	}
	if resumed.Status != StatusHealthy || resumed.MaintenanceMode {
		t.Fatalf("expected healthy service after resume, got %#v", resumed)
	}

	events, err := manager.ListEvents(context.Background())
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(events) < 2 || events[0].EventType != "service.resumed" || events[1].EventType != "service.maintenance.enabled" {
		t.Fatalf("unexpected recent events %#v", events[:2])
	}
}

func TestManagerSchedulerActions(t *testing.T) {
	current := time.Date(2026, 3, 26, 15, 0, 0, 0, time.UTC)
	manager := newManagerWithClock(version.Info{Version: "1.2.3"}, func() time.Time { return current })

	paused, err := manager.PauseScheduler(context.Background(), schedulerSyncID, "ops-user")
	if err != nil {
		t.Fatalf("PauseScheduler() error = %v", err)
	}
	if paused.Enabled || paused.NextRunAt != nil || paused.LastResult != ResultPaused {
		t.Fatalf("expected paused scheduler, got %#v", paused)
	}

	resumed, err := manager.ResumeScheduler(context.Background(), schedulerSyncID, "ops-user")
	if err != nil {
		t.Fatalf("ResumeScheduler() error = %v", err)
	}
	if !resumed.Enabled || resumed.NextRunAt == nil || resumed.LastResult != ResultOK {
		t.Fatalf("expected resumed scheduler, got %#v", resumed)
	}

	runOnce, err := manager.RunSchedulerOnce(context.Background(), schedulerSyncID, "ops-user")
	if err != nil {
		t.Fatalf("RunSchedulerOnce() error = %v", err)
	}
	if runOnce.LastRunAt == nil || runOnce.LastResult != ResultManualTriggered || runOnce.LastDurationMS == 0 {
		t.Fatalf("expected manual run result, got %#v", runOnce)
	}
}

func TestManagerReturnsNotFound(t *testing.T) {
	manager := newManagerWithClock(version.Info{Version: "1.2.3"}, time.Now)

	if _, err := manager.GetService(context.Background(), "missing"); err == nil || err.Error() == "" {
		t.Fatal("expected missing service error")
	}
	if _, err := manager.GetScheduler(context.Background(), "missing"); err == nil {
		t.Fatal("expected missing scheduler error")
	}
	if _, err := manager.ResumeService(context.Background(), "missing", "ops-user"); err == nil {
		t.Fatal("expected missing service error on resume")
	}
	if _, err := manager.ResumeScheduler(context.Background(), "missing", "ops-user"); err == nil {
		t.Fatal("expected missing scheduler error on resume")
	}
}

func TestManagerWrapsStoreNotFound(t *testing.T) {
	manager := newManagerWithClock(version.Info{Version: "1.2.3"}, time.Now)

	_, err := manager.GetService(context.Background(), "missing")
	if err == nil || err == store.ErrNotFound {
		t.Fatalf("expected wrapped not found error, got %v", err)
	}
}

func TestNewManagerAndListSchedulers(t *testing.T) {
	manager := NewManager(version.Info{Version: "2.0.0"})
	if manager == nil {
		t.Fatal("expected non-nil manager")
	}

	schedulers, err := manager.ListSchedulers(context.Background())
	if err != nil {
		t.Fatalf("ListSchedulers() error = %v", err)
	}
	if len(schedulers) != 3 {
		t.Fatalf("expected 3 schedulers, got %d", len(schedulers))
	}
	if schedulers[0].Name != "Control-plane sync" {
		t.Fatalf("expected sorted schedulers, got first %s", schedulers[0].Name)
	}
}

func TestOverviewStatusBranches(t *testing.T) {
	current := time.Date(2026, 3, 26, 15, 0, 0, 0, time.UTC)
	manager := newManagerWithClock(version.Info{Version: "1.2.3"}, func() time.Time { return current })

	manager.services[serviceExampleApp].activeStatus = StatusDown
	overview, err := manager.Overview(context.Background())
	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if overview.Status != StatusDown {
		t.Fatalf("expected down overview, got %s", overview.Status)
	}

	manager.services[serviceExampleApp].activeStatus = StatusHealthy
	manager.services[serviceClawbotServer].status.MaintenanceMode = true
	overview, err = manager.Overview(context.Background())
	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if overview.Status != StatusMaintenance {
		t.Fatalf("expected maintenance overview, got %s", overview.Status)
	}

	manager.services[serviceClawbotServer].status.MaintenanceMode = false
	overview, err = manager.Overview(context.Background())
	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if overview.Status != StatusHealthy {
		t.Fatalf("expected healthy overview, got %s", overview.Status)
	}
}

func TestListEventsRetentionAndFormattingHelpers(t *testing.T) {
	current := time.Date(2026, 3, 26, 15, 0, 0, 0, time.UTC)
	manager := newManagerWithClock(version.Info{Version: "1.2.3"}, func() time.Time { return current })

	for index := 0; index < 60; index++ {
		manager.mu.Lock()
		manager.recordEventLocked(serviceClawbotServer, "", "service.tick", SeverityInfo, "Periodic status update.")
		manager.mu.Unlock()
	}

	events, err := manager.ListEvents(context.Background())
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(events) != defaultRetentionLimit {
		t.Fatalf("expected %d retained events, got %d", defaultRetentionLimit, len(events))
	}
	if events[0].Message != "Periodic status update." {
		t.Fatalf("expected unwrapped message for empty actor, got %s", events[0].Message)
	}
}

func TestCopyDependencyStatusAndFormatEventMessage(t *testing.T) {
	empty := copyDependencyStatus(nil)
	if len(empty) != 0 {
		t.Fatalf("expected empty dependency map, got %#v", empty)
	}

	values := copyDependencyStatus(map[string]string{"postgres": StatusHealthy})
	values["postgres"] = StatusDown
	if updated := copyDependencyStatus(map[string]string{"postgres": StatusHealthy}); updated["postgres"] != StatusHealthy {
		t.Fatalf("expected fresh dependency copy, got %#v", updated)
	}

	if got := formatEventMessage("", "Service updated."); got != "Service updated." {
		t.Fatalf("unexpected message without actor: %s", got)
	}
	if got := formatEventMessage("ops-user", "Service updated."); got != "Service updated. (ops-user)" {
		t.Fatalf("unexpected message with actor: %s", got)
	}
}
