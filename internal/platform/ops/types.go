package ops

import "time"

const (
	StatusHealthy     = "healthy"
	StatusDegraded    = "degraded"
	StatusDown        = "down"
	StatusMaintenance = "maintenance"

	SeverityInfo  = "info"
	SeverityWarn  = "warn"
	SeverityError = "error"

	ResultOK                = "ok"
	ResultPaused            = "paused"
	ResultManualTriggered   = "manual_triggered"
	ResultWarning           = "warning"
	DefaultActivityPageSize = 25
)

type ServiceStatus struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	ServiceType      string            `json:"service_type"`
	Status           string            `json:"status"`
	Version          string            `json:"version"`
	UptimeSeconds    int64             `json:"uptime_seconds"`
	LastHeartbeatAt  time.Time         `json:"last_heartbeat_at"`
	MaintenanceMode  bool              `json:"maintenance_mode"`
	LastError        string            `json:"last_error"`
	DependencyStatus map[string]string `json:"dependency_status"`
}

type SchedulerStatus struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Enabled         bool       `json:"enabled"`
	IntervalSeconds int        `json:"interval_seconds"`
	LastRunAt       *time.Time `json:"last_run_at"`
	NextRunAt       *time.Time `json:"next_run_at"`
	LastResult      string     `json:"last_result"`
	LastDurationMS  int64      `json:"last_duration_ms"`
	LastError       string     `json:"last_error"`
}

type ActivityEvent struct {
	ID        string    `json:"id"`
	Time      time.Time `json:"time"`
	Source    string    `json:"source"`
	EventType string    `json:"event_type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
}

type Overview struct {
	Status              string    `json:"status"`
	ServicesTotal       int       `json:"services_total"`
	ServicesHealthy     int       `json:"services_healthy"`
	ServicesDegraded    int       `json:"services_degraded"`
	ServicesDown        int       `json:"services_down"`
	ServicesMaintenance int       `json:"services_maintenance"`
	SchedulersActive    int       `json:"schedulers_active"`
	SchedulersPaused    int       `json:"schedulers_paused"`
	RecentFailures      int       `json:"recent_failures"`
	LastUpdatedAt       time.Time `json:"last_updated_at"`
}
