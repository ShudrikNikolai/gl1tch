package credpg

import (
	"log/slog"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/storage/pg"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/credentials/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.Repository = (*CredentialsRepo)(nil)

type CredentialsRepo struct {
	log *slog.Logger
	pgx *pgxpool.Pool

	qCreate     string
	qByUsername string
	qUpdateHash string
	qDelete     string
}

func NewCredentialsPgRepo(client *pg.Client, log *slog.Logger) *CredentialsRepo {
	const table = "credentials"

	return &CredentialsRepo{
		pgx: client.Pool(),
		log: log,

		qCreate:     "INSERT INTO " + table + " (user_id, username, password_hash) VALUES ($1::uuid, $2, $3)",
		qByUsername: "SELECT user_id::text, username, password_hash, created_at FROM " + table + " WHERE lower(username) = lower($1)",
		qUpdateHash: "UPDATE " + table + " SET password_hash = $2 WHERE user_id = $1::uuid",
		qDelete:     "DELETE FROM " + table + " WHERE user_id = $1::uuid",
	}
}
