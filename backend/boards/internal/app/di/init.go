package di

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/adapters/grpc/users"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/storage/pg"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/storage/pg/migrations"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards/repository/boardpg"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts/repository/postpg"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads/repository/threadpg"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/transport/grpc_api"
	httpapi "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/transport/http_api"
	boarduc "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/boards"
	postuc "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/posts"
	threaduc "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/grpcx"
	_ "github.com/jackc/pgx/v5/stdlib" // регистрирует драйвер "pgx" для database/sql (нужен goose)
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

func (c *Container) initStorage(ctx context.Context) error {
	pgx, err := pg.New(ctx, c.Log, c.Cfg.Database.DatabaseURL)
	if err != nil {
		return err
	}
	c.Pgx = pgx

	return nil
}

// initMigration применяет миграции goose. Блокировка на уровне сессии не даёт нескольким репликам мигрировать одновременно.
func (c *Container) initMigration(ctx context.Context) error {
	db, err := sql.Open("pgx", c.Cfg.Database.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	locker, err := lock.NewPostgresSessionLocker(
		lock.WithLockTimeout(10, 30), // ждать до 300с (10с * 30 попыток)
	)
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
	c.boardRepo = boardpg.New(c.Pgx, c.Log)
	c.threadRepo = threadpg.New(c.Pgx, c.Log)
	c.postRepo = postpg.New(c.Pgx, c.Log)

	return nil
}

func (c *Container) initClients(_ context.Context) error {
	usersClient, err := users.New(c.Cfg.ServicesGRPCAddr.UsersGRPCAddr, c.Cfg.UsersCacheTTL)
	if err != nil {
		return err
	}
	c.Users = usersClient

	return nil
}

func (c *Container) initUseCases(_ context.Context) error {
	c.BoardUC = boarduc.New(c.boardRepo)

	c.ThreadUC = threaduc.New(threaduc.Deps{
		Tx:      c.Pgx,
		Boards:  c.boardRepo,
		Threads: c.threadRepo,
		Posts:   c.postRepo,
		Users:   c.Users,
		Log:     c.Log,
	})

	c.PostUC = postuc.New(postuc.Deps{
		Tx:      c.Pgx,
		Boards:  c.boardRepo,
		Threads: c.threadRepo,
		Posts:   c.postRepo,
		Users:   c.Users,
		Log:     c.Log,
	})

	return nil
}

func (c *Container) initServers() error {
	grpcSrv := grpcx.NewServer(c.Log)
	grpc_api.Register(grpcSrv, c.BoardUC, c.ThreadUC, c.PostUC)
	healthSrv := grpcx.RegisterHealth(grpcSrv)

	httpSrv := &http.Server{
		Handler:           httpapi.New(map[string]httpapi.Pinger{"postgres": c.Pgx}),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	c.Grpc = grpcSrv
	c.Http = httpSrv
	c.Health = healthSrv

	return nil
}
