package di

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/config"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/storage/pg"
	httpapi "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/transport/http_api"
	authcase "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/auth"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/cache/redis"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/grpcx"
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

	AuthCase *authcase.UseCase
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) (*Container, error) {
	c := &Container{
		Cfg: cfg,
		Log: log,
	}

	if err := c.initStorage(ctx); err != nil {
		log.Error("Failed to init storage", "error", err)
		return nil, err
	}

	if err := c.initRepo(ctx); err != nil {
		log.Error("Failed to init repo", "error", err)
		return nil, err
	}

	if err := c.initService(ctx); err != nil {
		log.Error("Failed to init service", "error", err)
		return nil, err
	}

	// todo init pg + redis -> собрать в отдельный пакет
	/*

		pool, err := pgrepo.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}

		rdb := redis.New(cfg.Redis, log)

		// Единая точка отката, пока контейнер не собран.
		closeAll := func() {
			rdb.Close()
			pool.Close()
		}

		if err := rdb.Ping(ctx); err != nil {
			closeAll()
			return nil, fmt.Errorf("redis: %w", err)
		}
	*/
	signer, err := jwtx.LoadSigner(
		cfg.Security.JWTPrivateKeyFile,
		cfg.Security.JWTIssuer,
		cfg.Security.AccessTTL,
	)

	if err != nil {
		// closeAll()
		return nil, err
	}

	// TODO для клиента так же
	/*
		users, err := newUsersClient(cfg, log)
		if err != nil {
			closeAll()
			return nil, err
		}

		repo := pgrepo.New(pool)

		svc, err := service.New(service.Deps{
			Repo:     {}, // repo,
			Sessions: {}, // session.NewStore(rdb, cfg.RefreshTTL, cfg.RefreshReuseGrace),
			Users:    {}, // users,
			Revoker:  revocation.NewPublisher(rdb, cfg.RevocationChannel),
			Hasher:   password.NewArgon2(cfg.Argon2),
			Signer:   signer,
			Log:      log,
		})
		if err != nil {
			closeAll()
			return nil, err
		}
	*/

	grpcSrv := grpcx.NewServer(log)
	// grpcapi.RegisterServer(grpcSrv, svc)
	healthSrv := grpcx.RegisterHealth(grpcSrv)
	httpSrv := &http.Server{
		Handler: httpapi.New(
			signer.JWKS,
			map[string]httpapi.Pinger{"postgres": nil, "redis": nil},
			//map[string]httpapi.Pinger{"postgres": pool, "redis": rdb},
		),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	c.Grpc = grpcSrv
	c.Http = httpSrv
	c.Health = healthSrv
	return c, nil
}

// Close освобождает ресурсы. Порядок важен: сначала redis (там сессии)
func (c *Container) Close() {
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
