package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/notifications/internal/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/notifications/internal/config"
	"github.com/ShudrikNikolai/gl1tch/backend/notifications/internal/consumer"
	"github.com/ShudrikNikolai/gl1tch/backend/notifications/internal/hub"
	"github.com/ShudrikNikolai/gl1tch/backend/notifications/internal/transport"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/jwtx"
)

type App struct {
	cfg    config.Config
	log    *slog.Logger
	boards *boards.Client
	http   *http.Server
	cancel context.CancelFunc
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) (*App, error) {
	boardsClient, err := boards.New(cfg.BoardsGRPCAddr)
	if err != nil {
		return nil, err
	}

	verifier := jwtx.NewVerifier(jwtx.VerifierConfig{
		JWKSURL:         cfg.JWKSURL,
		Issuer:          cfg.JWTIssuer,
		RefreshInterval: cfg.JWKSRefresh,
		MinRefreshGap:   cfg.JWKSMinRefresh,
	}, log)

	h := hub.New()
	consumerCtx, cancel := context.WithCancel(ctx)
	go consumer.Run(consumerCtx, cfg.Brokers(), cfg.KafkaTopic, cfg.KafkaGroup,
		consumer.NewHandler(h, log).Handle, log)

	httpSrv := &http.Server{
		Handler:           transport.New(verifier, boardsClient, h, log),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &App{cfg: cfg, log: log, boards: boardsClient, http: httpSrv, cancel: cancel}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.cancel()
	defer func() { _ = a.boards.Close() }()

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.HTTPPort))
	if err != nil {
		return err
	}

	a.log.Info("notifications started", slog.Int("http_port", a.cfg.HTTPPort))
	errCh := make(chan error, 1)
	go func() {
		if err := a.http.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err = <-errCh:
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.cfg.ShutdownTimeout)
	defer cancel()
	_ = a.http.Shutdown(shutdownCtx)

	return err
}
