package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	sharedConfig "github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/config"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/logger"
	"github.com/jcmturner/gokrb5/v8/config"
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
