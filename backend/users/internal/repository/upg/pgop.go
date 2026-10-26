package upg

import (
	"context"
	"errors"
	"fmt"

	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const pgUniqueViolation = "23505"
const userColumns = `user_id, username, bio, gender, roles, user_status, created_at, updated_at, deleted_at, banned_until`

func (u UserRepository) Create(ctx context.Context, userID, username string) error {
	_, err := u.pgx.Exec(ctx, "INSERT INTO users VALUES($1, $2)", userID, username)
	println(err)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return model.ErrUsernameTaken
		}

		return fmt.Errorf("users: create: %w", err)
	}

	return nil
}

func (u UserRepository) Me(ctx context.Context, userID string) (model.User, error) {
	var m model.User

	err := u.pgx.QueryRow(ctx, "SELECT * FROM users WHERE userID=$1", userID).Scan(&m)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return model.User{}, model.ErrUsernameTaken
		}

		return model.User{}, fmt.Errorf("credentials: me: %w", err)
	}

	return m, nil
}

func (u UserRepository) List(ctx context.Context, cursor *model.CursorRequest) ([]model.User, model.CursorResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserRepository) Get(ctx context.Context, userID string) (model.User, error) {
	rows, _ := u.pgx.Query(ctx, `
		SELECT user_id, username, bio, roles, user_status, gender,
		       created_at, updated_at, deleted_at, banned_until
		FROM users WHERE user_id = $1
	`, userID)

	defer rows.Close()
	m, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.User])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, model.ErrNotFound
		}
		return model.User{}, fmt.Errorf("users: get: %w", err)
	}
	return m, nil
}
