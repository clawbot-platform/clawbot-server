package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"clawbot-server/internal/config"
	"clawbot-server/internal/db"
	"clawbot-server/internal/http/handlers"
	"clawbot-server/internal/http/routes"
	"clawbot-server/internal/identityclient"
	"clawbot-server/internal/platform/audit"
	"clawbot-server/internal/platform/bots"
	"clawbot-server/internal/platform/ops"
	"clawbot-server/internal/platform/policies"
	"clawbot-server/internal/platform/runs"
	"clawbot-server/internal/platform/scheduler"
	"clawbot-server/internal/platform/store"
	"clawbot-server/internal/version"
	"clawbot-server/internal/watchlistidentity"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewLogger(level string, writer io.Writer) *slog.Logger {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slogLevel}))
}

func RunServer(ctx context.Context, cfg config.Server, logger *slog.Logger) error {
	if logger != nil {
		slog.SetDefault(logger)
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}
	defer pool.Close()

	if cfg.AutoMigrate {
		if err := db.ApplyAll(ctx, pool); err != nil {
			return fmt.Errorf("apply migrations: %w", err)
		}
	}

	pg := store.NewPostgres(pool)
	buildInfo := version.Current()
	services := buildServices(pg, buildInfo, cfg)
	identityRuntime := startIdentityEventRuntime(cfg, logger)
	defer identityRuntime.Close()

	router := routes.New(logger, routes.Services{
		System:              routes.NewSystemHandler(pg.Ping),
		Runs:                services.runs,
		Bots:                services.bots,
		Policies:            services.policies,
		Dashboard:           services.dashboard,
		Ops:                 services.ops,
		IdentityIntegration: handlers.NewIdentityIntegrationHandler(services.watchlistIdentity),
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return fmt.Errorf("listen and serve: %w", err)
	}
}

func MigrateUp(ctx context.Context, cfg config.Server, _ *slog.Logger) error {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}
	defer pool.Close()

	return db.ApplyAll(ctx, pool)
}

func MigrateDown(ctx context.Context, cfg config.Server, _ *slog.Logger) error {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}
	defer pool.Close()

	return db.DownOne(ctx, pool)
}

type appServices struct {
	runs              runs.Service
	bots              bots.Service
	policies          policies.Service
	dashboard         *store.DashboardReader
	ops               *ops.Manager
	watchlistIdentity *watchlistidentity.Service
}

func buildServices(pg *store.Postgres, buildInfo version.Info, cfg config.Server) appServices {
	auditRepo := audit.NewPostgresRepository()
	audits := audit.NewService(auditRepo)
	schedulerService := scheduler.NewPlaceholderService(audits)

	runsRepo := runs.NewPostgresRepository()
	botsRepo := bots.NewPostgresRepository()
	policiesRepo := policies.NewPostgresRepository()

	memoryClient := runs.MemoryClient(runs.NewNoopMemoryClient())
	if strings.TrimSpace(cfg.ClawmemBaseURL) != "" {
		memoryClient = runs.NewHTTPMemoryClient(cfg.ClawmemBaseURL, cfg.ClawmemTimeout)
	}

	inferenceClient := runs.InferenceClient(runs.NewNoopInferenceClient())
	if strings.TrimSpace(cfg.InferenceBaseURL) != "" {
		inferenceClient = runs.NewHTTPInferenceClient(cfg.InferenceBaseURL, cfg.InferenceTimeout)
	}

	identityClient := identityclient.New(cfg.IdentityBaseURL, cfg.IdentityTimeout, cfg.IdentityTenant)
	watchlistIntegration := watchlistidentity.NewService(identityClient)

	return appServices{
		runs: runs.NewManagerWithIntegrations(
			pg.Pool(),
			pg,
			runsRepo,
			audits,
			schedulerService,
			memoryClient,
			inferenceClient,
			runs.DependencyConfig{
				ClawmemBaseURL:               cfg.ClawmemBaseURL,
				InferenceBaseURL:             cfg.InferenceBaseURL,
				GuardrailTimeout:             cfg.GuardrailTimeout,
				HelperTimeout:                cfg.HelperTimeout,
				DisableLocalOllamaGuardrails: cfg.DisableLocalOllamaGuardrails,
				EnableCompactDualPayload:     cfg.EnableCompactDualPayload,
				PolicyBundleID:               cfg.PolicyBundleID,
				PolicyBundleVersion:          cfg.PolicyBundleVersion,
				Environment:                  cfg.AppEnv,
			},
		),
		bots:              bots.NewManager(pg.Pool(), pg, botsRepo, audits),
		policies:          policies.NewManager(pg.Pool(), pg, policiesRepo, audits),
		dashboard:         store.NewDashboardReader(pg.Pool()),
		ops:               ops.NewManager(buildInfo),
		watchlistIdentity: watchlistIntegration,
	}
}
