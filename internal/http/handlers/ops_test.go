package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"clawbot-server/internal/platform/ops"
	"clawbot-server/internal/platform/store"
	"clawbot-server/internal/version"

	"github.com/go-chi/chi/v5"
)

func TestOpsAPIEndpoints(t *testing.T) {
	handler := newTestOpsHandler()
	router := chi.NewRouter()
	router.Route("/api/v1/ops", func(r chi.Router) {
		r.Get("/overview", handler.Overview)
		r.Get("/services", handler.ListServices)
		r.Get("/services/{serviceID}", handler.GetService)
		r.Post("/services/{serviceID}/maintenance", handler.SetMaintenance)
		r.Post("/services/{serviceID}/resume", handler.ResumeService)
		r.Get("/schedulers", handler.ListSchedulers)
		r.Get("/schedulers/{schedulerID}", handler.GetScheduler)
		r.Post("/schedulers/{schedulerID}/pause", handler.PauseScheduler)
		r.Post("/schedulers/{schedulerID}/resume", handler.ResumeScheduler)
		r.Post("/schedulers/{schedulerID}/run-once", handler.RunSchedulerOnce)
		r.Get("/events", handler.ListEvents)
	})

	assertStatusCode(t, router, http.MethodGet, "/api/v1/ops/overview", http.StatusOK)
	assertStatusCode(t, router, http.MethodGet, "/api/v1/ops/services", http.StatusOK)
	assertStatusCode(t, router, http.MethodGet, "/api/v1/ops/services/clawbot-server", http.StatusOK)
	assertStatusCode(t, router, http.MethodGet, "/api/v1/ops/schedulers", http.StatusOK)
	assertStatusCode(t, router, http.MethodGet, "/api/v1/ops/schedulers/control-plane-sync", http.StatusOK)
	assertStatusCode(t, router, http.MethodGet, "/api/v1/ops/events", http.StatusOK)

	resp := performRequest(router, http.MethodPost, "/api/v1/ops/services/clawbot-server/maintenance", "")
	if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), "\"maintenance_mode\":true") {
		t.Fatalf("unexpected maintenance response %d %s", resp.Code, resp.Body.String())
	}

	resp = performRequest(router, http.MethodPost, "/api/v1/ops/services/clawbot-server/resume", "")
	if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), "\"maintenance_mode\":false") {
		t.Fatalf("unexpected resume response %d %s", resp.Code, resp.Body.String())
	}

	resp = performRequest(router, http.MethodPost, "/api/v1/ops/schedulers/control-plane-sync/pause", "")
	if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), "\"enabled\":false") {
		t.Fatalf("unexpected pause response %d %s", resp.Code, resp.Body.String())
	}

	resp = performRequest(router, http.MethodPost, "/api/v1/ops/schedulers/control-plane-sync/resume", "")
	if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), "\"enabled\":true") {
		t.Fatalf("unexpected scheduler resume response %d %s", resp.Code, resp.Body.String())
	}

	resp = performRequest(router, http.MethodPost, "/api/v1/ops/schedulers/control-plane-sync/run-once", "")
	if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), "\"last_result\":\"manual_triggered\"") {
		t.Fatalf("unexpected run-once response %d %s", resp.Code, resp.Body.String())
	}
}

func TestOpsAPIGetMissingReturnsNotFound(t *testing.T) {
	handler := newTestOpsHandler()
	router := chi.NewRouter()
	router.Get("/api/v1/ops/services/{serviceID}", handler.GetService)
	router.Get("/api/v1/ops/schedulers/{schedulerID}", handler.GetScheduler)

	resp := performRequest(router, http.MethodGet, "/api/v1/ops/services/missing", "")
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing service, got %d", resp.Code)
	}

	resp = performRequest(router, http.MethodGet, "/api/v1/ops/schedulers/missing", "")
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing scheduler, got %d", resp.Code)
	}
}

func TestOpsConsolePagesRender(t *testing.T) {
	handler := newTestOpsHandler()
	router := chi.NewRouter()
	router.Get("/ops", handler.OverviewPage)
	router.Get("/ops/services", handler.ServicesPage)
	router.Get("/ops/services/{serviceID}", handler.ServiceDetailPage)
	router.Get("/ops/schedulers", handler.SchedulersPage)
	router.Get("/ops/events", handler.EventsPage)

	assertBodyContains(t, router, http.MethodGet, "/ops", "Overview")
	assertBodyContains(t, router, http.MethodGet, "/ops/services", "Services / Clawbots")
	assertBodyContains(t, router, http.MethodGet, "/ops/services/clawbot-server", "Dependency status")
	assertBodyContains(t, router, http.MethodGet, "/ops/schedulers", "Schedulers / Jobs")
	assertBodyContains(t, router, http.MethodGet, "/ops/events", "Recent activity")
}

func TestOverviewPageCoversTrimAndIntermediateErrors(t *testing.T) {
	trimHandler := NewOpsHandler(stubOpsService{
		overview: ops.Overview{Status: ops.StatusHealthy},
		services: []ops.ServiceStatus{{ID: "svc-1", Name: "alpha"}},
		events: []ops.ActivityEvent{
			{ID: "1", Message: "event-1"},
			{ID: "2", Message: "event-2"},
			{ID: "3", Message: "event-3"},
			{ID: "4", Message: "event-4"},
			{ID: "5", Message: "event-5"},
			{ID: "6", Message: "event-6"},
			{ID: "7", Message: "event-7"},
		},
	})
	trimRouter := chi.NewRouter()
	trimRouter.Get("/ops", trimHandler.OverviewPage)
	resp := performRequest(trimRouter, http.MethodGet, "/ops", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 for trimmed overview page, got %d", resp.Code)
	}
	if strings.Contains(resp.Body.String(), "event-7") || !strings.Contains(resp.Body.String(), "event-6") {
		t.Fatalf("expected overview page to trim to six events, got %s", resp.Body.String())
	}

	listServicesErrHandler := NewOpsHandler(stubOpsService{
		overview:        ops.Overview{Status: ops.StatusHealthy},
		listServicesErr: errors.New("list services failed"),
	})
	listServicesErrRouter := chi.NewRouter()
	listServicesErrRouter.Get("/ops", listServicesErrHandler.OverviewPage)
	assertStatusCode(t, listServicesErrRouter, http.MethodGet, "/ops", http.StatusBadRequest)

	listEventsErrHandler := NewOpsHandler(stubOpsService{
		overview:      ops.Overview{Status: ops.StatusHealthy},
		services:      []ops.ServiceStatus{{ID: "svc-1", Name: "alpha"}},
		listEventsErr: errors.New("list events failed"),
	})
	listEventsErrRouter := chi.NewRouter()
	listEventsErrRouter.Get("/ops", listEventsErrHandler.OverviewPage)
	assertStatusCode(t, listEventsErrRouter, http.MethodGet, "/ops", http.StatusBadRequest)
}

func TestOpsConsolePageActionsRedirect(t *testing.T) {
	handler := newTestOpsHandler()
	router := chi.NewRouter()
	router.Post("/ops/services/{serviceID}/maintenance", handler.SetMaintenancePage)
	router.Post("/ops/services/{serviceID}/resume", handler.ResumeServicePage)
	router.Post("/ops/schedulers/{schedulerID}/pause", handler.PauseSchedulerPage)
	router.Post("/ops/schedulers/{schedulerID}/resume", handler.ResumeSchedulerPage)
	router.Post("/ops/schedulers/{schedulerID}/run-once", handler.RunSchedulerOncePage)

	req := httptest.NewRequest(http.MethodPost, "/ops/services/clawbot-server/maintenance", nil)
	req.Header.Set("Referer", "/ops/services")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusSeeOther || resp.Header().Get("Location") != "/ops/services" {
		t.Fatalf("unexpected maintenance redirect %d %s", resp.Code, resp.Header().Get("Location"))
	}

	assertRedirect(t, router, http.MethodPost, "/ops/services/clawbot-server/resume", "/ops/services")
	assertRedirect(t, router, http.MethodPost, "/ops/schedulers/control-plane-sync/pause", "/ops/schedulers")
	assertRedirect(t, router, http.MethodPost, "/ops/schedulers/control-plane-sync/resume", "/ops/schedulers")
	assertRedirect(t, router, http.MethodPost, "/ops/schedulers/control-plane-sync/run-once", "/ops/schedulers")
}

func TestOpsConsolePageActionErrors(t *testing.T) {
	handler := NewOpsHandler(stubOpsService{
		setMaintenanceErr:  errors.New("maintenance failed"),
		resumeServiceErr:   errors.New("resume failed"),
		pauseSchedulerErr:  errors.New("pause failed"),
		resumeSchedulerErr: errors.New("resume scheduler failed"),
		runOnceErr:         errors.New("run once failed"),
	})
	router := chi.NewRouter()
	router.Post("/ops/services/{serviceID}/maintenance", handler.SetMaintenancePage)
	router.Post("/ops/services/{serviceID}/resume", handler.ResumeServicePage)
	router.Post("/ops/schedulers/{schedulerID}/pause", handler.PauseSchedulerPage)
	router.Post("/ops/schedulers/{schedulerID}/resume", handler.ResumeSchedulerPage)
	router.Post("/ops/schedulers/{schedulerID}/run-once", handler.RunSchedulerOncePage)

	assertStatusCode(t, router, http.MethodPost, "/ops/services/clawbot-server/maintenance", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodPost, "/ops/services/clawbot-server/resume", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodPost, "/ops/schedulers/control-plane-sync/pause", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodPost, "/ops/schedulers/control-plane-sync/resume", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodPost, "/ops/schedulers/control-plane-sync/run-once", http.StatusBadRequest)
}

func TestOpsOverviewPayloadShape(t *testing.T) {
	handler := newTestOpsHandler()
	router := chi.NewRouter()
	router.Get("/api/v1/ops/overview", handler.Overview)

	resp := performRequest(router, http.MethodGet, "/api/v1/ops/overview", "")
	var payload struct {
		Data ops.Overview `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if payload.Data.ServicesTotal == 0 || payload.Data.Status == "" {
		t.Fatalf("unexpected payload %#v", payload)
	}
}

func TestOpsAPIErrorPaths(t *testing.T) {
	genericErr := errors.New("ops service unavailable")
	notFoundErr := store.ErrNotFound
	handler := NewOpsHandler(stubOpsService{
		overviewErr:        genericErr,
		listServicesErr:    genericErr,
		setMaintenanceErr:  genericErr,
		resumeServiceErr:   genericErr,
		listSchedulersErr:  genericErr,
		pauseSchedulerErr:  genericErr,
		resumeSchedulerErr: genericErr,
		runOnceErr:         genericErr,
		listEventsErr:      genericErr,
		getServiceErr:      notFoundErr,
		getSchedulerErr:    notFoundErr,
	})
	router := chi.NewRouter()
	router.Route("/api/v1/ops", func(r chi.Router) {
		r.Get("/overview", handler.Overview)
		r.Get("/services", handler.ListServices)
		r.Get("/services/{serviceID}", handler.GetService)
		r.Post("/services/{serviceID}/maintenance", handler.SetMaintenance)
		r.Post("/services/{serviceID}/resume", handler.ResumeService)
		r.Get("/schedulers", handler.ListSchedulers)
		r.Get("/schedulers/{schedulerID}", handler.GetScheduler)
		r.Post("/schedulers/{schedulerID}/pause", handler.PauseScheduler)
		r.Post("/schedulers/{schedulerID}/resume", handler.ResumeScheduler)
		r.Post("/schedulers/{schedulerID}/run-once", handler.RunSchedulerOnce)
		r.Get("/events", handler.ListEvents)
	})

	assertStatusCode(t, router, http.MethodGet, "/api/v1/ops/overview", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodGet, "/api/v1/ops/services", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodGet, "/api/v1/ops/services/missing", http.StatusNotFound)
	assertStatusCode(t, router, http.MethodPost, "/api/v1/ops/services/missing/maintenance", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodPost, "/api/v1/ops/services/missing/resume", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodGet, "/api/v1/ops/schedulers", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodGet, "/api/v1/ops/schedulers/missing", http.StatusNotFound)
	assertStatusCode(t, router, http.MethodPost, "/api/v1/ops/schedulers/missing/pause", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodPost, "/api/v1/ops/schedulers/missing/resume", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodPost, "/api/v1/ops/schedulers/missing/run-once", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodGet, "/api/v1/ops/events", http.StatusBadRequest)
}

func TestOpsConsolePageErrorPathsAndHelpers(t *testing.T) {
	genericErr := errors.New("page load failed")
	handler := NewOpsHandler(stubOpsService{
		overviewErr:       genericErr,
		listServicesErr:   genericErr,
		listSchedulersErr: genericErr,
		listEventsErr:     genericErr,
	})
	router := chi.NewRouter()
	router.Get("/ops", handler.OverviewPage)
	router.Get("/ops/services", handler.ServicesPage)
	router.Get("/ops/schedulers", handler.SchedulersPage)
	router.Get("/ops/events", handler.EventsPage)

	assertStatusCode(t, router, http.MethodGet, "/ops", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodGet, "/ops/services", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodGet, "/ops/schedulers", http.StatusBadRequest)
	assertStatusCode(t, router, http.MethodGet, "/ops/events", http.StatusBadRequest)

	notFoundHandler := NewOpsHandler(stubOpsService{
		overview:      ops.Overview{Status: ops.StatusHealthy},
		getServiceErr: store.ErrNotFound,
	})
	notFoundRouter := chi.NewRouter()
	notFoundRouter.Get("/ops/services/{serviceID}", notFoundHandler.ServiceDetailPage)
	assertStatusCode(t, notFoundRouter, http.MethodGet, "/ops/services/missing", http.StatusNotFound)

	req := httptest.NewRequest(http.MethodPost, "/ops/services/clawbot-server/maintenance", nil)
	resp := httptest.NewRecorder()
	redirectBack(resp, req, "/ops/services")
	if resp.Code != http.StatusSeeOther || resp.Header().Get("Location") != "/ops/services" {
		t.Fatalf("unexpected fallback redirect %d %s", resp.Code, resp.Header().Get("Location"))
	}

	if got := formatOpsTime(time.Time{}); got != "never" {
		t.Fatalf("expected never for zero time, got %s", got)
	}
	if got := formatUptime(5); got != "5s" {
		t.Fatalf("expected seconds uptime, got %s", got)
	}
	if got := statusClass("unexpected"); got != "status-degraded" {
		t.Fatalf("expected degraded class for unknown status, got %s", got)
	}
}

func TestRenderPageFailureReturnsServerError(t *testing.T) {
	handler := newTestOpsHandler()
	handler.templates = template.Must(template.New("broken").Parse(`{{ define "base" }}{{ .Content }}{{ end }}`))

	resp := httptest.NewRecorder()
	handler.renderPage(resp, "missing-template", opsPageData{Title: "broken"})

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.Code)
	}
}

func TestStatusAndUptimeHelpers(t *testing.T) {
	if got := statusClass(ops.StatusMaintenance); got != "status-maintenance" {
		t.Fatalf("unexpected maintenance class %s", got)
	}
	if got := statusClass(ops.StatusDown); got != "status-down" {
		t.Fatalf("unexpected down class %s", got)
	}
	if got := formatUptime(180); got != "3m" {
		t.Fatalf("expected minute uptime, got %s", got)
	}
	if got := formatUptime(3900); got != "1h 5m" {
		t.Fatalf("expected hour uptime, got %s", got)
	}
}

func newTestOpsHandler() *OpsHandler {
	service := ops.NewManager(version.Info{Version: "1.2.3"})
	return NewOpsHandler(service)
}

type stubOpsService struct {
	overview           ops.Overview
	overviewErr        error
	services           []ops.ServiceStatus
	listServicesErr    error
	service            ops.ServiceStatus
	getServiceErr      error
	setMaintenanceErr  error
	resumeServiceErr   error
	schedulers         []ops.SchedulerStatus
	listSchedulersErr  error
	scheduler          ops.SchedulerStatus
	getSchedulerErr    error
	pauseSchedulerErr  error
	resumeSchedulerErr error
	runOnceErr         error
	events             []ops.ActivityEvent
	listEventsErr      error
}

func (s stubOpsService) Overview(context.Context) (ops.Overview, error) {
	return s.overview, s.overviewErr
}
func (s stubOpsService) ListServices(context.Context) ([]ops.ServiceStatus, error) {
	return s.services, s.listServicesErr
}
func (s stubOpsService) GetService(context.Context, string) (ops.ServiceStatus, error) {
	return s.service, s.getServiceErr
}
func (s stubOpsService) SetMaintenance(context.Context, string, string) (ops.ServiceStatus, error) {
	return s.service, s.setMaintenanceErr
}
func (s stubOpsService) ResumeService(context.Context, string, string) (ops.ServiceStatus, error) {
	return s.service, s.resumeServiceErr
}
func (s stubOpsService) ListSchedulers(context.Context) ([]ops.SchedulerStatus, error) {
	return s.schedulers, s.listSchedulersErr
}
func (s stubOpsService) GetScheduler(context.Context, string) (ops.SchedulerStatus, error) {
	return s.scheduler, s.getSchedulerErr
}
func (s stubOpsService) PauseScheduler(context.Context, string, string) (ops.SchedulerStatus, error) {
	return s.scheduler, s.pauseSchedulerErr
}
func (s stubOpsService) ResumeScheduler(context.Context, string, string) (ops.SchedulerStatus, error) {
	return s.scheduler, s.resumeSchedulerErr
}
func (s stubOpsService) RunSchedulerOnce(context.Context, string, string) (ops.SchedulerStatus, error) {
	return s.scheduler, s.runOnceErr
}
func (s stubOpsService) ListEvents(context.Context) ([]ops.ActivityEvent, error) {
	return s.events, s.listEventsErr
}

func performRequest(handler http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp
}

func assertStatusCode(t *testing.T, handler http.Handler, method string, path string, want int) {
	t.Helper()
	resp := performRequest(handler, method, path, "")
	if resp.Code != want {
		t.Fatalf("%s %s: expected %d, got %d body=%s", method, path, want, resp.Code, resp.Body.String())
	}
}

func assertBodyContains(t *testing.T, handler http.Handler, method string, path string, needle string) {
	t.Helper()
	resp := performRequest(handler, method, path, "")
	if resp.Code != http.StatusOK {
		t.Fatalf("%s %s: expected 200, got %d", method, path, resp.Code)
	}
	if !strings.Contains(resp.Body.String(), needle) {
		t.Fatalf("%s %s: expected body to contain %q, got %s", method, path, needle, resp.Body.String())
	}
}

func assertRedirect(t *testing.T, handler http.Handler, method string, path string, want string) {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusSeeOther || resp.Header().Get("Location") != want {
		t.Fatalf("%s %s: expected redirect to %s, got %d %s", method, path, want, resp.Code, resp.Header().Get("Location"))
	}
}
