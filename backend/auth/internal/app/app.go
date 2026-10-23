package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/app/di"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/config"
)

type App struct {
	di *di.Container
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) (*App, error) {
	container, err := di.New(ctx, cfg, log)
	if err != nil {
		return nil, err
	}

	return &App{
		di: container,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.di.Close()
	log := a.di.Log
	cfg := a.di.Cfg

	httpLn, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.HTTPPort))
	if err != nil {
		return err
	}
	grpcLn, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		_ = httpLn.Close()
		return err
	}

	log.Info("auth started", slog.Int("http_port", cfg.HTTPPort), slog.Int("grpc_port", cfg.GRPCPort))

	return a.di.Serve(ctx, httpLn, grpcLn)
}
