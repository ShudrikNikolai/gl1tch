package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	sharedConfig "github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/config"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/healthcheck"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/logger"
	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/app"
	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/infrastructure/config"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg := sharedConfig.MustLoad[config.Config]()
	healthcheck.Maybe(os.Args, cfg.HTTPPort)
	log := logger.New("users", cfg.Env, cfg.Logger())

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx, cfg, log)
	if err != nil {
		log.Error("failed to create app", slog.Any("error", err))
		return 1
	}

	if err = application.Run(ctx); err != nil {
		log.Error("users stopped with error", slog.Any("error", err))
		return 1
	}

	log.Info("users down")

	return 0
}
