package di

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/adapters/grpc/users"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/config"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/storage/pg"
	credrepo "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/credentials/repository"
	credsvc "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/credentials/service"
	rtrepo "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens/repository"
	rtsvc "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens/service"
	authcase "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/auth"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/cache/redis"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/jwtx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Container struct {
	Cfg config.Config
	Log *slog.Logger

	Grpc   *grpc.Server
	Health *health.Server
	Http   *http.Server

	Rdb *redis.Client
	Pgx *pg.Client

	Signer *jwtx.Signer
	Users  *users.Client

	AuthCase *authcase.UseCase

	credRepo credrepo.Repository
	rtRepo   rtrepo.Repository
	credSvc  *credsvc.Service
	rtSvc    *rtsvc.Service
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) (*Container, error) {
	c := &Container{
		Cfg: cfg,
		Log: log,
	}

	steps := []struct {
		name string
		run  func() error
	}{
		{"storage", func() error { return c.initStorage(ctx) }},
		{"migrate", func() error { return c.initMigration(ctx) }},
		{"repo", func() error { return c.initRepo(ctx) }},
		{"clients", func() error { return c.initClients(ctx) }},
		{"services", func() error { return c.initServices(ctx) }},
		{"usecases", func() error { return c.initUseCases(ctx) }},
		{"servers", func() error { return c.initServers() }},
	}
	for _, s := range steps {
		if err := s.run(); err != nil {
			log.Error("Failed to init "+s.name, "error", err)
			c.Close()

			return nil, err
		}
	}

	return c, nil
}

// Close освобождает ресурсы. Порядок важен: сначала клиент users, потом redis (там сессии), потом postgres.
// Серверы останавливает Serve.
func (c *Container) Close() {
	if c.Users != nil {
		_ = c.Users.Close()
	}
	if c.Rdb != nil {
		_ = c.Rdb.Close()
	}
	if c.Pgx != nil {
		c.Pgx.Close()
	}
}

func (c *Container) Serve(ctx context.Context, httpLn, grpcLn net.Listener) error {
	errChs := make(chan error, 2)

	go func() {
		errChs <- c.Grpc.Serve(grpcLn)
	}()
	go func() {
		if err := c.Http.Serve(httpLn); !errors.Is(err, http.ErrServerClosed) {
			errChs <- err
		}
	}()

	c.Health.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	var serveErr error
	select {
	case <-ctx.Done():
	case serveErr = <-errChs:
		if serveErr != nil {
			c.Log.Error("server stopped unexpectedly", "err", serveErr)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), c.Cfg.ShutdownTimeout)
	defer cancel()

	c.Health.Shutdown()

	doneCh := make(chan struct{})
	go func() {
		c.Grpc.GracefulStop()
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-shutdownCtx.Done():
		c.Log.Warn("grpc graceful stop timed out, forcing")
		c.Grpc.Stop()
	}

	if err := c.Http.Shutdown(shutdownCtx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			c.Log.Warn("http shutdown timed out, forcing close")
			_ = c.Http.Close()
		} else {
			c.Log.Error("http shutdown", "err", err)
			if serveErr == nil {
				serveErr = err
			}
		}
	}

	return serveErr
}
