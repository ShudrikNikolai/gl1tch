package upg

import (
	"log/slog"

	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/infrastructure/storage/pg"
	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.Repository = (*UserRepository)(nil)

type UserRepository struct {
	log       *slog.Logger
	pgx       *pgxpool.Pool
	tableName string
}

func NewUsersPgRepo(client *pg.Client, log *slog.Logger) *UserRepository {
	const table = "users"

	return &UserRepository{
		pgx:       client.Pool(),
		log:       log,
		tableName: table,
	}
}
