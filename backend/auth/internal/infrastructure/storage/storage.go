package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pg interface {
	Pool() *pgxpool.Pool
	Ping(ctx context.Context) error
	Close()
}
