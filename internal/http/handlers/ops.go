package handlers

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"path"
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

	data := opsPageData{
		Title:       "Overview",
		CurrentPath: opsOverviewPath,
		Overview:    overview,
		Services:    services,
		Events:      events,
	}
	h.renderPage(w, "overview-content", data)
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

	data := opsPageData{
		Title:       "Services",
		CurrentPath: opsServicesPath,
		Overview:    overview,
		Services:    services,
	}
	h.renderPage(w, "services-content", data)
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

	data := opsPageData{
		Title:       fmt.Sprintf("Service: %s", service.Name),
		CurrentPath: opsServicesPath,
		Overview:    overview,
		Service:     service,
	}
	h.renderPage(w, "service-detail-content", data)
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

	data := opsPageData{
		Title:       "Schedulers",
		CurrentPath: opsSchedulersPath,
		Overview:    overview,
		Schedulers:  schedulers,
	}
	h.renderPage(w, "schedulers-content", data)
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

	data := opsPageData{
		Title:       "Recent activity",
		CurrentPath: opsEventsPath,
		Overview:    overview,
		Events:      events,
	}
	h.renderPage(w, "events-content", data)
}

func (h *OpsHandler) SetMaintenancePage(w http.ResponseWriter, r *http.Request) {
	h.handleServicePageAction(w, r, opsServicesPath, func(ctx context.Context, id string, actor string) error {
		_, err := h.service.SetMaintenance(ctx, id, actor)
		return err
	})
}

func (h *OpsHandler) ResumeServicePage(w http.ResponseWriter, r *http.Request) {
	h.handleServicePageAction(w, r, opsServicesPath, func(ctx context.Context, id string, actor string) error {
		_, err := h.service.ResumeService(ctx, id, actor)
		return err
	})
}

func (h *OpsHandler) PauseSchedulerPage(w http.ResponseWriter, r *http.Request) {
	h.handleSchedulerPageAction(w, r, opsSchedulersPath, func(ctx context.Context, id string, actor string) error {
		_, err := h.service.PauseScheduler(ctx, id, actor)
		return err
	})
}

func (h *OpsHandler) ResumeSchedulerPage(w http.ResponseWriter, r *http.Request) {
	h.handleSchedulerPageAction(w, r, opsSchedulersPath, func(ctx context.Context, id string, actor string) error {
		_, err := h.service.ResumeScheduler(ctx, id, actor)
		return err
	})
}

func (h *OpsHandler) RunSchedulerOncePage(w http.ResponseWriter, r *http.Request) {
	h.handleSchedulerPageAction(w, r, opsSchedulersPath, func(ctx context.Context, id string, actor string) error {
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

func redirectBack(w http.ResponseWriter, r *http.Request, fallback string) {
	http.Redirect(w, r, safeOpsRedirectTarget(r, fallback), http.StatusSeeOther)
}

func safeOpsRedirectTarget(r *http.Request, fallback string) string {
	normalizedFallback := normalizeOpsRedirectTarget(fallback)
	if normalizedFallback == "" {
		normalizedFallback = opsOverviewPath
	}

	referer := strings.TrimSpace(r.Referer())
	if referer == "" {
		return normalizedFallback
	}

	refURL, err := url.Parse(referer)
	if err != nil {
		return normalizedFallback
	}

	// Only allow same-origin referers when the host is present.
	if refURL.Host != "" && !sameOriginHost(refURL.Host, requestHost(r)) {
		return normalizedFallback
	}

	targetPath := normalizeOpsRedirectTarget(refURL.EscapedPath())
	if targetPath == "" {
		return normalizedFallback
	}

	if refURL.RawQuery != "" {
		return targetPath + "?" + refURL.RawQuery
	}
	return targetPath
}

func normalizeOpsRedirectTarget(target string) string {
	target = strings.TrimSpace(target)
	if target == "" {
		return ""
	}

	parsed, err := url.Parse(target)
	if err != nil || parsed.IsAbs() {
		return ""
	}

	cleanPath := path.Clean(parsed.EscapedPath())
	if cleanPath == "." || cleanPath == "" {
		cleanPath = "/"
	}

	if !strings.HasPrefix(cleanPath, opsOverviewPath) {
		return ""
	}

	switch {
	case cleanPath == opsOverviewPath:
		return opsOverviewPath
	case cleanPath == opsServicesPath:
		return opsServicesPath
	case strings.HasPrefix(cleanPath, opsServicesPath+"/"):
		return cleanPath
	case cleanPath == opsSchedulersPath:
		return opsSchedulersPath
	case strings.HasPrefix(cleanPath, opsSchedulersPath+"/"):
		return cleanPath
	case cleanPath == opsEventsPath:
		return opsEventsPath
	default:
		return ""
	}
}

func requestHost(r *http.Request) string {
	if forwardedHost := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); forwardedHost != "" {
		if strings.Contains(forwardedHost, ",") {
			parts := strings.Split(forwardedHost, ",")
			return strings.TrimSpace(parts[0])
		}
		return forwardedHost
	}
	if r.Host != "" {
		return r.Host
	}
	return r.URL.Host
}

func sameOriginHost(a, b string) bool {
	aHost := canonicalHost(a)
	bHost := canonicalHost(b)
	return aHost != "" && bHost != "" && strings.EqualFold(aHost, bHost)
}

func canonicalHost(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	host, port, err := net.SplitHostPort(value)
	if err == nil {
		if port == "" || port == "80" || port == "443" {
			return strings.ToLower(host)
		}
		return strings.ToLower(net.JoinHostPort(host, port))
	}

	return strings.ToLower(value)
}

func (h *OpsHandler) handleServicePageAction(
	w http.ResponseWriter,
	r *http.Request,
	fallback string,
	action func(context.Context, string, string) error,
) {
	if err := action(r.Context(), chi.URLParam(r, "serviceID"), actorFromRequest(r)); err != nil {
		writeServiceError(w, err)
		return
	}
	redirectBack(w, r, fallback)
}

func (h *OpsHandler) handleSchedulerPageAction(
	w http.ResponseWriter,
	r *http.Request,
	fallback string,
	action func(context.Context, string, string) error,
) {
	if err := action(r.Context(), chi.URLParam(r, "schedulerID"), actorFromRequest(r)); err != nil {
		writeServiceError(w, err)
		return
	}
	redirectBack(w, r, fallback)
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
