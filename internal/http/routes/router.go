package routes

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"clawbot-server/internal/http/handlers"
	mw "clawbot-server/internal/http/middleware"
	"clawbot-server/internal/platform/bots"
	"clawbot-server/internal/platform/policies"
	"clawbot-server/internal/platform/runs"
	"clawbot-server/internal/version"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Services struct {
	System    *handlers.SystemHandler
	Runs      runs.Service
	Bots      bots.Service
	Policies  policies.Service
	Dashboard handlers.DashboardService
}

func New(logger *slog.Logger, services Services) http.Handler {
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(mw.CaptureRequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.Timeout(30 * time.Second))
	router.Use(mw.RequestLogger(logger))

	router.Get("/healthz", services.System.Health)
	router.Get("/readyz", services.System.Ready)
	router.Get("/version", services.System.Version)

	router.Route("/api/v1", func(r chi.Router) {
		runsHandler := handlers.NewRunsHandler(services.Runs)
		botsHandler := handlers.NewBotsHandler(services.Bots)
		policiesHandler := handlers.NewPoliciesHandler(services.Policies)
		dashboardHandler := handlers.NewDashboardHandler(services.Dashboard)

		r.Get("/dashboard/summary", dashboardHandler.Summary)

		r.Route("/runs", func(r chi.Router) {
			r.Get("/", runsHandler.List)
			r.Post("/", runsHandler.Create)
			r.Get("/{runID}", runsHandler.Get)
			r.Patch("/{runID}", runsHandler.Update)
		})

		r.Route("/bots", func(r chi.Router) {
			r.Get("/", botsHandler.List)
			r.Post("/", botsHandler.Create)
			r.Get("/{botID}", botsHandler.Get)
			r.Patch("/{botID}", botsHandler.Update)
		})

		r.Route("/policies", func(r chi.Router) {
			r.Get("/", policiesHandler.List)
			r.Post("/", policiesHandler.Create)
			r.Get("/{policyID}", policiesHandler.Get)
			r.Patch("/{policyID}", policiesHandler.Update)
		})
	})

	return router
}

type readinessFunc func(context.Context) error

func NewSystemHandler(readiness readinessFunc) *handlers.SystemHandler {
	return handlers.NewSystemHandler(readiness, version.Current())
}
