package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ShudrikNikolai/gl1tch/backend/notifications/internal/app"
	"github.com/ShudrikNikolai/gl1tch/backend/notifications/internal/config"
	sharedConfig "github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/config"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/healthcheck"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/logger"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/otelx"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg := sharedConfig.MustLoad[config.Config]()
	healthcheck.Maybe(os.Args, cfg.HTTPPort)
	log := logger.New("notifications", cfg.Env, cfg.Logger())

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if cfg.OTLPEndpoint != "" {
		shutdown, err := otelx.Init(ctx, otelx.Config{
			ServiceName:  "notifications",
			Environment:  cfg.Env,
			OTLPEndpoint: cfg.OTLPEndpoint,
		})
		if err != nil {
			log.Warn("otel disabled", slog.Any("error", err))
		} else {
			defer func() { _ = shutdown(context.Background()) }()
		}
	}

	application, err := app.New(ctx, cfg, log)
	if err != nil {
		log.Error("failed to create app", slog.Any("error", err))
		return 1
	}

	if err = application.Run(ctx); err != nil {
		log.Error("notifications stopped with error", slog.Any("error", err))
		return 1
	}

	log.Info("notifications down")

	return 0
}
