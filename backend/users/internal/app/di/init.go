package di

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/grpcx"
	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/infrastructure/storage/pg"
	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/infrastructure/storage/pg/migrations"
	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/service/users"
	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/transport/grpc_api"
	httpapi "github.com/ShudrikNikolai/gl1tch/backend/users/internal/transport/http_api"
	"google.golang.org/grpc/reflection"

	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/repository/upg"
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

	return nil
}

func (c *Container) initMigration(ctx context.Context) error {
	db, err := sql.Open("pgx", c.Cfg.Database.DatabaseURL)
	println("Starting migration")
	if err != nil {
		return err
	}
	defer db.Close()

	locker, err := lock.NewPostgresSessionLocker(lock.WithLockTimeout(10, 30))
	if err != nil {
		return err
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		migrations.FS,
		goose.WithSessionLocker(locker),
		goose.WithTableName("users_schema_migrations"),
	)
	if err != nil {
		return err
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	println("Starting migration \n", results)

	c.Log.Info("migrations applied", "count", len(results))

	return nil
}

func (c *Container) initRepo(_ context.Context) error {
	c.UsersRepo = upg.NewUsersPgRepo(c.Pgx, c.Log)

	return nil
}

func (c *Container) initServices(_ context.Context) error {
	c.UsersSrv = users.New(users.Deps{
		UsersRepo: c.UsersRepo,
		Log:       c.Log,
	})

	return nil
}

func (c *Container) initServers() error {
	grpcSrv := grpcx.NewServer(c.Log)

	grpc_api.RegisterUsersGrpcServer(grpcSrv, c.UsersSrv)
	//TODO не забыть потом вырубить
	reflection.Register(grpcSrv)

	healthSrv := grpcx.RegisterHealth(grpcSrv)

	httpSrv := &http.Server{
		Handler: httpapi.New(
			map[string]httpapi.Pinger{"postgres": c.Pgx},
		),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	c.Grpc = grpcSrv
	c.Http = httpSrv
	c.Health = healthSrv

	return nil
}
