package di

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/adapters/grpc/users"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/password"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/storage/pg"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/storage/pg/migrations"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/credentials/repository/credpg"
	credsvc "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/credentials/service"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens/repository/refredis"
	rtsvc "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens/service"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/transport/grpc_api"
	httpapi "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/transport/http_api"
	authcase "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/auth"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/cache/redis"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/grpcx"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/jwtx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

const startupPingTimeout = 5 * time.Second

func (c *Container) initStorage(ctx context.Context) error {
	pgx, err := pg.New(ctx, c.Log, c.Cfg.Database)
	if err != nil {
		return err
	}
	c.Pgx = pgx

	c.Rdb = redis.New(c.Cfg.Redis, c.Log)

	pingCtx, cancel := context.WithTimeout(ctx, startupPingTimeout)
	defer cancel()
	if err = c.Rdb.Ping(pingCtx); err != nil {
		return fmt.Errorf("redis: %w", err)
	}

	return nil
}

func (c *Container) initMigration(ctx context.Context) error {
	db, err := sql.Open("pgx", c.Cfg.Database.DatabaseURL)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	locker, err := lock.NewPostgresSessionLocker(lock.WithLockTimeout(10, 30))
	if err != nil {
		return err
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		migrations.FS,
		goose.WithSessionLocker(locker),
		goose.WithTableName("schema_migrations"),
	)
	if err != nil {
		return err
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	c.Log.Info("migrations applied", "count", len(results))

	return nil
}

func (c *Container) initRepo(_ context.Context) error {
	c.credRepo = credpg.NewCredentialsPgRepo(c.Pgx, c.Log)
	c.rtRepo = refredis.New(c.Rdb, c.Log)

	return nil
}

func (c *Container) initClients(_ context.Context) error {
	usersClient, err := users.New(c.Cfg.ServicesGRPCAddr.UsersGRPCAddr)
	if err != nil {
		return err
	}
	c.Users = usersClient

	return nil
}

func (c *Container) initServices(_ context.Context) error {
	sec := c.Cfg.Security

	signer, err := jwtx.LoadSigner(sec.JWTPrivateKeyFile, sec.JWTIssuer, sec.AccessTTL)
	if err != nil {
		return err
	}
	c.Signer = signer

	c.credSvc, err = credsvc.New(c.credRepo, password.NewArgon2(c.Cfg.Argon2), c.Log)
	if err != nil {
		return err
	}

	revoker := rtsvc.NewPublisher(c.Rdb, sec.RevocationChannel)
	c.rtSvc = rtsvc.New(c.rtRepo, revoker, sec.RefreshTTL, sec.RefreshReuseGrace, c.Log)

	return nil
}

func (c *Container) initUseCases(_ context.Context) error {
	c.AuthCase = authcase.New(authcase.Deps{
		Credentials: c.credSvc,
		Refresh:     c.rtSvc,
		Users:       c.Users,
		Signer:      c.Signer,
		Log:         c.Log,
	})

	return nil
}

func (c *Container) initServers() error {
	sec := c.Cfg.Security

	grpcSrv := grpcx.NewServer(c.Log)

	grpc_api.RegisterAuthGrpcServer(grpcSrv, c.AuthCase, grpc_api.CookieConfig{
		Name:   sec.CookieName,
		Path:   sec.CookiePath,
		Secure: sec.CookieSecure,
		MaxAge: sec.RefreshTTL,
	})

	healthSrv := grpcx.RegisterHealth(grpcSrv)

	httpSrv := &http.Server{
		Handler: httpapi.New(
			c.Signer.JWKS,
			map[string]httpapi.Pinger{"postgres": c.Pgx, "redis": c.Rdb},
		),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	c.Grpc = grpcSrv
	c.Http = httpSrv
	c.Health = healthSrv

	return nil
}
