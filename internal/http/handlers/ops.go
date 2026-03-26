package handlers

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"clawbot-server/internal/platform/ops"

	"github.com/go-chi/chi/v5"
)

//go:embed templates/ops/*.gohtml
var opsTemplateFS embed.FS

const (
	opsOverviewPath   = "/ops"
	opsServicesPath   = "/ops/services"
	opsSchedulersPath = "/ops/schedulers"
	opsEventsPath     = "/ops/events"

	returnOverview      = "overview"
	returnServices      = "services"
	returnServiceDetail = "service_detail"
	returnSchedulers    = "schedulers"
	returnEvents        = "events"
)

type OpsService interface {
	Overview(context.Context) (ops.Overview, error)
	ListServices(context.Context) ([]ops.ServiceStatus, error)
	GetService(context.Context, string) (ops.ServiceStatus, error)
	SetMaintenance(context.Context, string, string) (ops.ServiceStatus, error)
	ResumeService(context.Context, string, string) (ops.ServiceStatus, error)
	ListSchedulers(context.Context) ([]ops.SchedulerStatus, error)
	GetScheduler(context.Context, string) (ops.SchedulerStatus, error)
	PauseScheduler(context.Context, string, string) (ops.SchedulerStatus, error)
	ResumeScheduler(context.Context, string, string) (ops.SchedulerStatus, error)
	RunSchedulerOnce(context.Context, string, string) (ops.SchedulerStatus, error)
	ListEvents(context.Context) ([]ops.ActivityEvent, error)
}

type OpsHandler struct {
	service   OpsService
	templates *template.Template
}

type opsPageData struct {
	Title           string
	CurrentPath     string
	Overview        ops.Overview
	Services        []ops.ServiceStatus
	Service         ops.ServiceStatus
	Schedulers      []ops.SchedulerStatus
	Scheduler       ops.SchedulerStatus
	Events          []ops.ActivityEvent
	ContentTemplate string
}

type dependencyEntry struct {
	Name   string
	Status string
}

func NewOpsHandler(service OpsService) *OpsHandler {
	funcs := template.FuncMap{
		"formatTime":      formatOpsTime,
		"formatMaybeTime": formatOptionalOpsTime,
		"formatUptime":    formatUptime,
		"statusClass":     statusClass,
		"dependencyPairs": dependencyPairs,
	}

	templates := template.Must(template.New("ops").Funcs(funcs).ParseFS(
		opsTemplateFS,
		"templates/ops/base.gohtml",
		"templates/ops/overview.gohtml",
		"templates/ops/services.gohtml",
		"templates/ops/service_detail.gohtml",
		"templates/ops/schedulers.gohtml",
		"templates/ops/events.gohtml",
	))

	return &OpsHandler{
		service:   service,
		templates: templates,
	}
}
func (h *OpsHandler) Overview(w http.ResponseWriter, r *http.Request) {
	overview, err := h.service.Overview(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": overview})
}

func (h *OpsHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	services, err := h.service.ListServices(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": services})
}

func (h *OpsHandler) GetService(w http.ResponseWriter, r *http.Request) {
	service, err := h.service.GetService(r.Context(), chi.URLParam(r, "serviceID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": service})
}

func (h *OpsHandler) SetMaintenance(w http.ResponseWriter, r *http.Request) {
	service, err := h.service.SetMaintenance(r.Context(), chi.URLParam(r, "serviceID"), actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": service})
}

func (h *OpsHandler) ResumeService(w http.ResponseWriter, r *http.Request) {
	service, err := h.service.ResumeService(r.Context(), chi.URLParam(r, "serviceID"), actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": service})
}

func (h *OpsHandler) ListSchedulers(w http.ResponseWriter, r *http.Request) {
	schedulers, err := h.service.ListSchedulers(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": schedulers})
}

func (h *OpsHandler) GetScheduler(w http.ResponseWriter, r *http.Request) {
	scheduler, err := h.service.GetScheduler(r.Context(), chi.URLParam(r, "schedulerID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": scheduler})
}

func (h *OpsHandler) PauseScheduler(w http.ResponseWriter, r *http.Request) {
	scheduler, err := h.service.PauseScheduler(r.Context(), chi.URLParam(r, "schedulerID"), actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": scheduler})
}

func (h *OpsHandler) ResumeScheduler(w http.ResponseWriter, r *http.Request) {
	scheduler, err := h.service.ResumeScheduler(r.Context(), chi.URLParam(r, "schedulerID"), actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": scheduler})
}

func (h *OpsHandler) RunSchedulerOnce(w http.ResponseWriter, r *http.Request) {
	scheduler, err := h.service.RunSchedulerOnce(r.Context(), chi.URLParam(r, "schedulerID"), actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": scheduler})
}

func (h *OpsHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.service.ListEvents(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": events})
}

func (h *OpsHandler) OverviewPage(w http.ResponseWriter, r *http.Request) {
	overview, err := h.service.Overview(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	services, err := h.service.ListServices(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	events, err := h.service.ListEvents(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if len(events) > 6 {
		events = events[:6]
	}

	h.renderPage(w, "overview-content", opsPageData{
		Title:       "Overview",
		CurrentPath: opsOverviewPath,
		Overview:    overview,
		Services:    services,
		Events:      events,
	})
}

func (h *OpsHandler) ServicesPage(w http.ResponseWriter, r *http.Request) {
	overview, err := h.service.Overview(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	services, err := h.service.ListServices(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	h.renderPage(w, "services-content", opsPageData{
		Title:       "Services",
		CurrentPath: opsServicesPath,
		Overview:    overview,
		Services:    services,
	})
}

func (h *OpsHandler) ServiceDetailPage(w http.ResponseWriter, r *http.Request) {
	overview, err := h.service.Overview(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	service, err := h.service.GetService(r.Context(), chi.URLParam(r, "serviceID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	h.renderPage(w, "service-detail-content", opsPageData{
		Title:       fmt.Sprintf("Service: %s", service.Name),
		CurrentPath: opsServicesPath,
		Overview:    overview,
		Service:     service,
	})
}

func (h *OpsHandler) SchedulersPage(w http.ResponseWriter, r *http.Request) {
	overview, err := h.service.Overview(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	schedulers, err := h.service.ListSchedulers(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	h.renderPage(w, "schedulers-content", opsPageData{
		Title:       "Schedulers",
		CurrentPath: opsSchedulersPath,
		Overview:    overview,
		Schedulers:  schedulers,
	})
}

func (h *OpsHandler) EventsPage(w http.ResponseWriter, r *http.Request) {
	overview, err := h.service.Overview(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	events, err := h.service.ListEvents(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	h.renderPage(w, "events-content", opsPageData{
		Title:       "Recent activity",
		CurrentPath: opsEventsPath,
		Overview:    overview,
		Events:      events,
	})
}
func (h *OpsHandler) handleServicePageAction(
	w http.ResponseWriter,
	r *http.Request,
	action func(context.Context, string, string) error,
) {
	serviceID := chi.URLParam(r, "serviceID")
	if err := action(r.Context(), serviceID, actorFromRequest(r)); err != nil {
		writeServiceError(w, err)
		return
	}
	redirectToReturnTarget(w, r, buildReturnTarget(r.FormValue("return_to"), serviceID))
}

func (h *OpsHandler) SetMaintenancePage(w http.ResponseWriter, r *http.Request) {
	h.handleServicePageAction(w, r, func(ctx context.Context, id string, actor string) error {
		_, err := h.service.SetMaintenance(ctx, id, actor)
		return err
	})
}

func (h *OpsHandler) ResumeServicePage(w http.ResponseWriter, r *http.Request) {
	h.handleServicePageAction(w, r, func(ctx context.Context, id string, actor string) error {
		_, err := h.service.ResumeService(ctx, id, actor)
		return err
	})
}

func (h *OpsHandler) PauseSchedulerPage(w http.ResponseWriter, r *http.Request) {
	h.handleSchedulerPageAction(w, r, func(ctx context.Context, id string, actor string) error {
		_, err := h.service.PauseScheduler(ctx, id, actor)
		return err
	})
}

func (h *OpsHandler) ResumeSchedulerPage(w http.ResponseWriter, r *http.Request) {
	h.handleSchedulerPageAction(w, r, func(ctx context.Context, id string, actor string) error {
		_, err := h.service.ResumeScheduler(ctx, id, actor)
		return err
	})
}

func (h *OpsHandler) RunSchedulerOncePage(w http.ResponseWriter, r *http.Request) {
	h.handleSchedulerPageAction(w, r, func(ctx context.Context, id string, actor string) error {
		_, err := h.service.RunSchedulerOnce(ctx, id, actor)
		return err
	})
}

func (h *OpsHandler) renderPage(w http.ResponseWriter, contentTemplate string, data opsPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data.ContentTemplate = contentTemplate
	if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
		writeError(w, http.StatusInternalServerError, "render_error", err.Error())
	}
}

func (h *OpsHandler) handleSchedulerPageAction(
	w http.ResponseWriter,
	r *http.Request,
	action func(context.Context, string, string) error,
) {
	schedulerID := chi.URLParam(r, "schedulerID")
	if err := action(r.Context(), schedulerID, actorFromRequest(r)); err != nil {
		writeServiceError(w, err)
		return
	}
	redirectToReturnTarget(w, r, buildReturnTarget(r.FormValue("return_to"), ""))
}

func redirectToReturnTarget(w http.ResponseWriter, r *http.Request, target string) {
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func buildReturnTarget(returnTo string, serviceID string) string {
	switch strings.TrimSpace(returnTo) {
	case returnOverview:
		return opsOverviewPath
	case returnServices:
		return opsServicesPath
	case returnServiceDetail:
		if serviceID != "" {
			return opsServicesPath + "/" + url.PathEscape(serviceID)
		}
		return opsServicesPath
	case returnSchedulers:
		return opsSchedulersPath
	case returnEvents:
		return opsEventsPath
	default:
		return opsOverviewPath
	}
}

func dependencyPairs(values map[string]string) []dependencyEntry {
	pairs := make([]dependencyEntry, 0, len(values))
	for name, status := range values {
		pairs = append(pairs, dependencyEntry{Name: name, Status: status})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Name < pairs[j].Name
	})
	return pairs
}

func formatOpsTime(value time.Time) string {
	if value.IsZero() {
		return "never"
	}
	return value.Local().Format("2006-01-02 15:04:05")
}

func formatOptionalOpsTime(value *time.Time) string {
	if value == nil {
		return "not scheduled"
	}
	return formatOpsTime(*value)
}

func formatUptime(seconds int64) string {
	duration := time.Duration(seconds) * time.Second
	if duration < time.Minute {
		return duration.String()
	}
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func statusClass(status string) string {
	switch status {
	case ops.StatusHealthy, ops.SeverityInfo:
		return "status-healthy"
	case ops.StatusMaintenance:
		return "status-maintenance"
	case ops.SeverityWarn:
		return "status-degraded"
	case ops.StatusDown, ops.SeverityError:
		return "status-down"
	default:
		return "status-degraded"
	}
}
