package ops

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"clawbot-server/internal/platform/store"
	"clawbot-server/internal/version"
)

const (
	defaultServiceTypeControlPlane = "control-plane"
	defaultServiceTypeMemory       = "memory"
	defaultServiceTypeApplication  = "application"

	serviceClawbotServer  = "clawbot-server"
	serviceClawmem        = "clawmem"
	serviceExampleApp     = "downstream-app"
	schedulerHeartbeatID  = "service-heartbeat-scan"
	schedulerSyncID       = "control-plane-sync"
	schedulerReconcileID  = "maintenance-reconciliation"
	defaultRunDurationMS  = int64(175)
	defaultRetentionLimit = 50
)

type Service interface {
	Overview(context.Context) (Overview, error)
	ListServices(context.Context) ([]ServiceStatus, error)
	GetService(context.Context, string) (ServiceStatus, error)
	SetMaintenance(context.Context, string, string) (ServiceStatus, error)
	ResumeService(context.Context, string, string) (ServiceStatus, error)
	ListSchedulers(context.Context) ([]SchedulerStatus, error)
	GetScheduler(context.Context, string) (SchedulerStatus, error)
	PauseScheduler(context.Context, string, string) (SchedulerStatus, error)
	ResumeScheduler(context.Context, string, string) (SchedulerStatus, error)
	RunSchedulerOnce(context.Context, string, string) (SchedulerStatus, error)
	ListEvents(context.Context) ([]ActivityEvent, error)
}

type Manager struct {
	mu         sync.RWMutex
	now        func() time.Time
	build      version.Info
	nextEvent  int
	services   map[string]*serviceRecord
	schedulers map[string]*schedulerRecord
	events     []ActivityEvent
}

type serviceRecord struct {
	status       ServiceStatus
	startedAt    time.Time
	activeStatus string
}

type schedulerRecord struct {
	status SchedulerStatus
}

func NewManager(build version.Info) *Manager {
	return newManagerWithClock(build, time.Now)
}

func newManagerWithClock(build version.Info, now func() time.Time) *Manager {
	current := now()
	manager := &Manager{
		now:        now,
		build:      build,
		services:   seedServices(build, current),
		schedulers: seedSchedulers(current),
		events:     seedEvents(current),
		nextEvent:  len(seedEvents(current)),
	}
	return manager
}

func (m *Manager) Overview(_ context.Context) (Overview, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var overview Overview
	overview.LastUpdatedAt = m.now()

	for _, record := range m.services {
		overview.ServicesTotal++
		switch serviceStatus(record) {
		case StatusHealthy:
			overview.ServicesHealthy++
		case StatusDegraded:
			overview.ServicesDegraded++
		case StatusDown:
			overview.ServicesDown++
		case StatusMaintenance:
			overview.ServicesMaintenance++
		}

		if record.status.LastError != "" {
			overview.RecentFailures++
		}
	}

	for _, record := range m.schedulers {
		if record.status.Enabled {
			overview.SchedulersActive++
		} else {
			overview.SchedulersPaused++
		}
		if record.status.LastError != "" {
			overview.RecentFailures++
		}
	}

	overview.Status = deriveOverallStatus(overview)
	return overview, nil
}

func (m *Manager) ListServices(_ context.Context) ([]ServiceStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make([]ServiceStatus, 0, len(m.services))
	for _, record := range m.services {
		services = append(services, snapshotService(record, m.now()))
	}
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})
	return services, nil
}

func (m *Manager) GetService(_ context.Context, id string) (ServiceStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	record, ok := m.services[id]
	if !ok {
		return ServiceStatus{}, fmt.Errorf("service %q: %w", id, store.ErrNotFound)
	}

	return snapshotService(record, m.now()), nil
}

func (m *Manager) SetMaintenance(_ context.Context, id string, actor string) (ServiceStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.setMaintenanceLocked(id, actor, true)
}

func (m *Manager) ResumeService(_ context.Context, id string, actor string) (ServiceStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.setMaintenanceLocked(id, actor, false)
}

func (m *Manager) ListSchedulers(_ context.Context) ([]SchedulerStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	schedulers := make([]SchedulerStatus, 0, len(m.schedulers))
	for _, record := range m.schedulers {
		schedulers = append(schedulers, snapshotScheduler(record))
	}
	sort.Slice(schedulers, func(i, j int) bool {
		return schedulers[i].Name < schedulers[j].Name
	})
	return schedulers, nil
}

func (m *Manager) GetScheduler(_ context.Context, id string) (SchedulerStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	record, ok := m.schedulers[id]
	if !ok {
		return SchedulerStatus{}, fmt.Errorf("scheduler %q: %w", id, store.ErrNotFound)
	}

	return snapshotScheduler(record), nil
}

func (m *Manager) PauseScheduler(_ context.Context, id string, actor string) (SchedulerStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, ok := m.schedulers[id]
	if !ok {
		return SchedulerStatus{}, fmt.Errorf("scheduler %q: %w", id, store.ErrNotFound)
	}

	record.status.Enabled = false
	record.status.NextRunAt = nil
	record.status.LastResult = ResultPaused
	record.status.LastError = ""
	m.recordEventLocked(record.status.ID, actor, "scheduler.paused", SeverityWarn, fmt.Sprintf("Paused scheduler %s.", record.status.Name))
	return snapshotScheduler(record), nil
}

func (m *Manager) ResumeScheduler(_ context.Context, id string, actor string) (SchedulerStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, ok := m.schedulers[id]
	if !ok {
		return SchedulerStatus{}, fmt.Errorf("scheduler %q: %w", id, store.ErrNotFound)
	}

	now := m.now()
	record.status.Enabled = true
	record.status.NextRunAt = timePointer(now.Add(time.Duration(record.status.IntervalSeconds) * time.Second))
	record.status.LastResult = ResultOK
	record.status.LastError = ""
	m.recordEventLocked(record.status.ID, actor, "scheduler.resumed", SeverityInfo, fmt.Sprintf("Resumed scheduler %s.", record.status.Name))
	return snapshotScheduler(record), nil
}

func (m *Manager) RunSchedulerOnce(_ context.Context, id string, actor string) (SchedulerStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, ok := m.schedulers[id]
	if !ok {
		return SchedulerStatus{}, fmt.Errorf("scheduler %q: %w", id, store.ErrNotFound)
	}

	now := m.now()
	record.status.LastRunAt = timePointer(now)
	record.status.LastResult = ResultManualTriggered
	record.status.LastDurationMS = defaultRunDurationMS
	record.status.LastError = ""
	if record.status.Enabled {
		record.status.NextRunAt = timePointer(now.Add(time.Duration(record.status.IntervalSeconds) * time.Second))
	}
	m.recordEventLocked(record.status.ID, actor, "scheduler.run_once", SeverityInfo, fmt.Sprintf("Triggered scheduler %s for one immediate run.", record.status.Name))
	return snapshotScheduler(record), nil
}

func (m *Manager) ListEvents(_ context.Context) ([]ActivityEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := append([]ActivityEvent(nil), m.events...)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Time.After(events[j].Time)
	})
	if len(events) > defaultRetentionLimit {
		events = events[:defaultRetentionLimit]
	}
	return events, nil
}

func (m *Manager) setMaintenanceLocked(id string, actor string, enabled bool) (ServiceStatus, error) {
	record, ok := m.services[id]
	if !ok {
		return ServiceStatus{}, fmt.Errorf("service %q: %w", id, store.ErrNotFound)
	}

	record.status.MaintenanceMode = enabled
	record.status.LastHeartbeatAt = m.now()
	eventType := "service.resumed"
	severity := SeverityInfo
	message := fmt.Sprintf("Resumed service %s from maintenance mode.", record.status.Name)
	if enabled {
		eventType = "service.maintenance.enabled"
		severity = SeverityWarn
		message = fmt.Sprintf("Placed service %s into maintenance mode.", record.status.Name)
	}

	m.recordEventLocked(record.status.ID, actor, eventType, severity, message)
	return snapshotService(record, m.now()), nil
}

func (m *Manager) recordEventLocked(source string, actor string, eventType string, severity string, message string) {
	m.nextEvent++
	entry := ActivityEvent{
		ID:        fmt.Sprintf("evt-%03d", m.nextEvent),
		Time:      m.now(),
		Source:    source,
		EventType: eventType,
		Severity:  severity,
		Message:   formatEventMessage(actor, message),
	}
	m.events = append([]ActivityEvent{entry}, m.events...)
	if len(m.events) > defaultRetentionLimit {
		m.events = m.events[:defaultRetentionLimit]
	}
}

func snapshotService(record *serviceRecord, now time.Time) ServiceStatus {
	status := record.status
	status.Status = serviceStatus(record)
	status.UptimeSeconds = int64(now.Sub(record.startedAt).Seconds())
	status.DependencyStatus = copyDependencyStatus(status.DependencyStatus)
	return status
}

func snapshotScheduler(record *schedulerRecord) SchedulerStatus {
	status := record.status
	if status.LastRunAt != nil {
		lastRun := *status.LastRunAt
		status.LastRunAt = &lastRun
	}
	if status.NextRunAt != nil {
		nextRun := *status.NextRunAt
		status.NextRunAt = &nextRun
	}
	return status
}

func serviceStatus(record *serviceRecord) string {
	if record.status.MaintenanceMode {
		return StatusMaintenance
	}
	return record.activeStatus
}

func deriveOverallStatus(overview Overview) string {
	switch {
	case overview.ServicesDown > 0:
		return StatusDown
	case overview.ServicesDegraded > 0:
		return StatusDegraded
	case overview.ServicesMaintenance > 0:
		return StatusMaintenance
	default:
		return StatusHealthy
	}
}

func seedServices(build version.Info, now time.Time) map[string]*serviceRecord {
	return map[string]*serviceRecord{
		serviceClawbotServer: newServiceRecord(ServiceStatus{
			ID:              serviceClawbotServer,
			Name:            "clawbot-server",
			ServiceType:     defaultServiceTypeControlPlane,
			Status:          StatusHealthy,
			Version:         build.Version,
			LastHeartbeatAt: now.Add(-15 * time.Second),
			DependencyStatus: map[string]string{
				"postgres": StatusHealthy,
				"redis":    StatusHealthy,
				"nats":     StatusHealthy,
			},
		}, now.Add(-3*time.Hour-12*time.Minute)),
		serviceClawmem: newServiceRecord(ServiceStatus{
			ID:              serviceClawmem,
			Name:            "clawmem",
			ServiceType:     defaultServiceTypeMemory,
			Status:          StatusHealthy,
			Version:         "v1",
			LastHeartbeatAt: now.Add(-25 * time.Second),
			DependencyStatus: map[string]string{
				"local-store": StatusHealthy,
			},
		}, now.Add(-2*time.Hour-5*time.Minute)),
		serviceExampleApp: newServiceRecord(ServiceStatus{
			ID:              serviceExampleApp,
			Name:            "downstream-app",
			ServiceType:     defaultServiceTypeApplication,
			Status:          StatusDegraded,
			Version:         "v1",
			LastHeartbeatAt: now.Add(-3 * time.Minute),
			LastError:       "Heartbeat latency exceeded the warning threshold.",
			DependencyStatus: map[string]string{
				"clawbot-server": StatusHealthy,
				"clawmem":        StatusHealthy,
			},
		}, now.Add(-95*time.Minute)),
	}
}

func seedSchedulers(now time.Time) map[string]*schedulerRecord {
	return map[string]*schedulerRecord{
		schedulerHeartbeatID: {
			status: SchedulerStatus{
				ID:              schedulerHeartbeatID,
				Name:            "Service heartbeat scan",
				Enabled:         true,
				IntervalSeconds: 60,
				LastRunAt:       timePointer(now.Add(-35 * time.Second)),
				NextRunAt:       timePointer(now.Add(25 * time.Second)),
				LastResult:      ResultWarning,
				LastDurationMS:  82,
			},
		},
		schedulerSyncID: {
			status: SchedulerStatus{
				ID:              schedulerSyncID,
				Name:            "Control-plane sync",
				Enabled:         true,
				IntervalSeconds: 300,
				LastRunAt:       timePointer(now.Add(-2 * time.Minute)),
				NextRunAt:       timePointer(now.Add(3 * time.Minute)),
				LastResult:      ResultOK,
				LastDurationMS:  140,
			},
		},
		schedulerReconcileID: {
			status: SchedulerStatus{
				ID:              schedulerReconcileID,
				Name:            "Maintenance reconciliation",
				Enabled:         false,
				IntervalSeconds: 900,
				LastRunAt:       timePointer(now.Add(-45 * time.Minute)),
				LastResult:      ResultPaused,
				LastDurationMS:  0,
				LastError:       "Paused by operator for maintenance review.",
			},
		},
	}
}

func seedEvents(now time.Time) []ActivityEvent {
	return []ActivityEvent{
		{
			ID:        "evt-003",
			Time:      now.Add(-5 * time.Minute),
			Source:    serviceExampleApp,
			EventType: "service.degraded",
			Severity:  SeverityWarn,
			Message:   "downstream-app reported delayed heartbeats and entered a degraded state.",
		},
		{
			ID:        "evt-002",
			Time:      now.Add(-12 * time.Minute),
			Source:    schedulerHeartbeatID,
			EventType: "scheduler.completed",
			Severity:  SeverityInfo,
			Message:   "Service heartbeat scan completed with one warning.",
		},
		{
			ID:        "evt-001",
			Time:      now.Add(-18 * time.Minute),
			Source:    serviceClawbotServer,
			EventType: "service.started",
			Severity:  SeverityInfo,
			Message:   "clawbot-server started and published its initial heartbeat.",
		},
	}
}

func newServiceRecord(status ServiceStatus, startedAt time.Time) *serviceRecord {
	return &serviceRecord{
		status:       status,
		startedAt:    startedAt,
		activeStatus: status.Status,
	}
}

func copyDependencyStatus(source map[string]string) map[string]string {
	if len(source) == 0 {
		return map[string]string{}
	}
	copied := make(map[string]string, len(source))
	for key, value := range source {
		copied[key] = value
	}
	return copied
}

func formatEventMessage(actor string, message string) string {
	if actor == "" {
		return message
	}
	return fmt.Sprintf("%s (%s)", message, actor)
}

func timePointer(value time.Time) *time.Time {
	return &value
}
