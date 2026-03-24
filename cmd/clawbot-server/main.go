package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"clawbot-server/internal/app"
	"clawbot-server/internal/config"
)

func main() {
	cfg, err := config.LoadServerFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load server config: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "clawbot-server: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Server, args []string) error {
	logger := app.NewLogger(cfg.LogLevel, os.Stdout)

	if len(args) == 0 || args[0] == "serve" {
		logger.Info("starting clawbot-server", slog.String("address", cfg.HTTPAddress))
		return app.RunServer(ctx, cfg, logger)
	}

	if args[0] != "migrate" {
		return fmt.Errorf("unknown command %q", args[0])
	}

	if len(args) < 2 {
		return fmt.Errorf("migrate requires a subcommand: up or down")
	}

	switch args[1] {
	case "up":
		return app.MigrateUp(ctx, cfg, logger)
	case "down":
		return app.MigrateDown(ctx, cfg, logger)
	default:
		return fmt.Errorf("unknown migrate subcommand %q", args[1])
	}
}
