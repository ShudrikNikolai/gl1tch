package di

import (
	"context"
	"net/http"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/storage/pg"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/transport/grpc_api"
	httpapi "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/transport/http_api"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/cache/redis"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/grpcx"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/jwtx"
)

func (c *Container) initStorage(ctx context.Context) error {
	pgx, err := pg.New(ctx, c.Log, c.Cfg.Database)
	if err != nil {
		return err
	}
	c.Pgx = pgx

	rdb := redis.New(c.Cfg.Redis, c.Log)
	println(rdb)
	c.Rdb = rdb

	return nil
}

func (c *Container) initRepo(ctx context.Context) error {
	return nil
}

func (c *Container) initService(ctx context.Context) error {
	return nil
}

func (c *Container) initServer() error {
	cfgSecurity := c.Cfg.Security
	signer, err := jwtx.LoadSigner(
		cfgSecurity.JWTPrivateKeyFile,
		cfgSecurity.JWTIssuer,
		cfgSecurity.AccessTTL,
	)

	if err != nil {
		return err
	}

	grpcSrv := grpcx.NewServer(c.Log)
	grpc_api.RegisterAuthGrpcServer(grpcSrv, c.AuthCase)

	healthSrv := grpcx.RegisterHealth(grpcSrv)

	httpSrv := &http.Server{
		Handler: httpapi.New(
			signer.JWKS,
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
