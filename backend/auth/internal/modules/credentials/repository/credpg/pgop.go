package credpg

import (
	"context"
	"errors"
	"fmt"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/credentials"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/credentials/repository/converter"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/credentials/repository/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const pgUniqueViolation = "23505"

func (r *CredentialsRepo) Create(ctx context.Context, c credentials.Credential) error {
	m := converter.CredentialSrvToModel(c)

	if _, err := r.pgx.Exec(ctx, r.qCreate, m.UserID, m.Username, m.PasswordHash); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return credentials.ErrUsernameTaken
		}

		return fmt.Errorf("credentials: create: %w", err)
	}

	return nil
}

func (r *CredentialsRepo) GetByUsername(ctx context.Context, username string) (credentials.Credential, error) {
	var m model.Credential

	err := r.pgx.QueryRow(ctx, r.qByUsername, username).Scan(&m.UserID, &m.Username, &m.PasswordHash, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return credentials.Credential{}, credentials.ErrNotFound
		}

		return credentials.Credential{}, fmt.Errorf("credentials: get by username: %w", err)
	}

	return converter.CredentialModelToSrv(m), nil
}

func (r *CredentialsRepo) UpdateHash(ctx context.Context, userID, passwordHash string) error {
	if _, err := r.pgx.Exec(ctx, r.qUpdateHash, userID, passwordHash); err != nil {
		return fmt.Errorf("credentials: update hash: %w", err)
	}

	return nil
}

func (r *CredentialsRepo) Delete(ctx context.Context, userID string) error {
	if _, err := r.pgx.Exec(ctx, r.qDelete, userID); err != nil {
		return fmt.Errorf("credentials: delete: %w", err)
	}

	return nil
}
