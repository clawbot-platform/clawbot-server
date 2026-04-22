package routes

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"clawbot-server/internal/http/handlers"
	mw "clawbot-server/internal/http/middleware"
	"clawbot-server/internal/platform/bots"
	"clawbot-server/internal/platform/ops"
	"clawbot-server/internal/platform/policies"
	"clawbot-server/internal/platform/runs"
	"clawbot-server/internal/version"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Services struct {
	System *handlers.SystemHandler

	Runs     runs.Service
	Bots     bots.Service
	Policies policies.Service

	Dashboard handlers.DashboardService

	Ops ops.Service

	IdentityIntegration *handlers.IdentityIntegrationHandler
	WatchlistReview     *handlers.WatchlistReviewHandler
}

func New(logger *slog.Logger, services Services) http.Handler {
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(mw.CaptureRequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.Timeout(3 * time.Minute))
	router.Use(mw.RequestLogger(logger))

	router.Get("/healthz", services.System.Health)
	router.Get("/readyz", services.System.Ready)
	router.Get("/version", services.System.Version)

	opsHandler := handlers.NewOpsHandler(services.Ops)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ops", http.StatusSeeOther)
	})

	router.Route("/ops", func(r chi.Router) {
		r.Get("/", opsHandler.OverviewPage)
		r.Get("/services", opsHandler.ServicesPage)
		r.Get("/services/{serviceID}", opsHandler.ServiceDetailPage)
		r.Post("/services/{serviceID}/maintenance", opsHandler.SetMaintenancePage)
		r.Post("/services/{serviceID}/resume", opsHandler.ResumeServicePage)
		r.Get("/schedulers", opsHandler.SchedulersPage)
		r.Post("/schedulers/{schedulerID}/pause", opsHandler.PauseSchedulerPage)
		r.Post("/schedulers/{schedulerID}/resume", opsHandler.ResumeSchedulerPage)
		r.Post("/schedulers/{schedulerID}/run-once", opsHandler.RunSchedulerOncePage)
		r.Get("/events", opsHandler.EventsPage)
	})

	router.Route("/api/v1", func(r chi.Router) {
		runsHandler := handlers.NewRunsHandler(services.Runs)
		botsHandler := handlers.NewBotsHandler(services.Bots)
		policiesHandler := handlers.NewPoliciesHandler(services.Policies)
		dashboardHandler := handlers.NewDashboardHandler(services.Dashboard)

		r.Get("/dashboard/summary", dashboardHandler.Summary)

		r.Route("/ops", func(r chi.Router) {
			r.Get("/overview", opsHandler.Overview)
			r.Get("/services", opsHandler.ListServices)
			r.Get("/services/{serviceID}", opsHandler.GetService)
			r.Post("/services/{serviceID}/maintenance", opsHandler.SetMaintenance)
			r.Post("/services/{serviceID}/resume", opsHandler.ResumeService)
			r.Get("/schedulers", opsHandler.ListSchedulers)
			r.Get("/schedulers/{schedulerID}", opsHandler.GetScheduler)
			r.Post("/schedulers/{schedulerID}/pause", opsHandler.PauseScheduler)
			r.Post("/schedulers/{schedulerID}/resume", opsHandler.ResumeScheduler)
			r.Post("/schedulers/{schedulerID}/run-once", opsHandler.RunSchedulerOnce)
			r.Get("/events", opsHandler.ListEvents)
		})

		r.Route("/runs", func(r chi.Router) {
			r.Get("/", runsHandler.List)
			r.Post("/", runsHandler.Create)
			r.Get("/{runID}", runsHandler.Get)
			r.Patch("/{runID}", runsHandler.Update)
			r.Post("/{runID}/approve", runsHandler.ApproveRun)
			r.Post("/{runID}/reject", runsHandler.RejectRun)
			r.Post("/{runID}/override", runsHandler.OverrideRun)
			r.Post("/{runID}/defer", runsHandler.DeferRun)
			r.Post("/{runID}/start", runsHandler.StartRun)
			r.Get("/{runID}/artifacts", runsHandler.ListArtifacts)
			r.Post("/{runID}/artifacts", runsHandler.AttachArtifact)
			r.Post("/{runID}/cycles", runsHandler.CreateCycle)
			r.Get("/{runID}/cycles/{cycleID}", runsHandler.GetCycle)
			r.Patch("/{runID}/cycles/{cycleID}", runsHandler.UpdateCycle)
			r.Post("/{runID}/cycles/{cycleID}/execute", runsHandler.ExecuteCycleRun)
			r.Get("/{runID}/comparison", runsHandler.GetComparison)
			r.Post("/{runID}/comparison", runsHandler.UpsertComparison)
		})

		r.Route("/model-profiles", func(r chi.Router) {
			r.Post("/", runsHandler.RegisterModelProfile)
			r.Get("/{modelProfileID}", runsHandler.GetModelProfile)
		})

		r.Route("/control-plane", func(r chi.Router) {
			r.Get("/dependencies", runsHandler.DependencyHealth)
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

		if services.IdentityIntegration != nil {
			r.Route("/identity", func(r chi.Router) {
				r.Post("/compare", services.IdentityIntegration.Compare)
				r.Post("/watchlist/ofac/screenings", services.IdentityIntegration.ScreenOFAC)
			})
		}

		if services.WatchlistReview != nil {
			r.Route("/watchlist-review", func(r chi.Router) {
				r.Post("/reviews", services.WatchlistReview.Create)
			})
		}
	})

	return router
}

type readinessFunc func(context.Context) error

func NewSystemHandler(readiness readinessFunc) *handlers.SystemHandler {
	return handlers.NewSystemHandler(readiness, version.Current())
}
