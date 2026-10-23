package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/app"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/config"
	sharedConfig "github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/config"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/logger"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	cfg := sharedConfig.MustLoad[config.Config]()
	log := logger.New("auth", cfg.Env, cfg.Logger())

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx, cfg, log)
	if err != nil {
		log.Error("Failed to create app: %v", err)
		os.Exit(1)
	}

	if err = application.Run(ctx); err != nil {
		log.Error("auth stopped with error", slog.Any("error", err))
		os.Exit(1)
	}

	log.Info("auth down -_-")
}
