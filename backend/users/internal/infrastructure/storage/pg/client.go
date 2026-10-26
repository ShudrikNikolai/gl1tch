package pg

import (
	"context"
	"log/slog"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/infrastructure/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	log  *slog.Logger
	cfg  config.Database
	pool *pgxpool.Pool
}

func New(ctx context.Context, log *slog.Logger, cfg config.Database) (*Client, error) {
	cfgPgx, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Error("Unable to parse config DatabaseURL ", "error", err)
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfgPgx)
	if err != nil {
		log.Error("Unable to connect to database", "error", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = pool.Ping(pingCtx); err != nil {
		pool.Close()
		log.Error("Unable to ping database", "error", err)
		return nil, err
	}

	return &Client{
		log:  log,
		cfg:  cfg,
		pool: pool,
	}, nil
}

func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
}

func (c *Client) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}

func (c *Client) Close() {
	c.pool.Close()
}
