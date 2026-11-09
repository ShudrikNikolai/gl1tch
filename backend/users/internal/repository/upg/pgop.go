package upg

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const pgUniqueViolation = "23505"

func (u UserRepository) Create(ctx context.Context, userID, username string) error {
	_, err := u.pgx.Exec(ctx, "INSERT INTO users (user_id, username) VALUES ($1::uuid, $2)", userID, username)
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
	return u.Get(ctx, userID)
}

func (u UserRepository) List(ctx context.Context, cursor *model.CursorRequest) ([]model.User, model.CursorResponse, error) {
	size := 20
	if cursor != nil && cursor.PageSize > 0 {
		size = int(cursor.PageSize)
	}
	if size > 100 {
		size = 100
	}

	offset := 0
	if cursor != nil && cursor.PageToken != "" {
		n, err := strconv.Atoi(cursor.PageToken)
		if err != nil || n < 0 {
			return nil, model.CursorResponse{}, apperr.InvalidArgument("INVALID_PAGE_TOKEN", "invalid page token")
		}
		offset = n
	}

	rows, err := u.pgx.Query(ctx, `
		SELECT user_id, username, bio, roles, user_status, gender,
		       created_at, updated_at, deleted_at, banned_until
		FROM users
		ORDER BY created_at, user_id
		LIMIT $1 OFFSET $2
	`, size+1, offset)
	if err != nil {
		return nil, model.CursorResponse{}, fmt.Errorf("users: list: %w", err)
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		return nil, model.CursorResponse{}, fmt.Errorf("users: list scan: %w", err)
	}

	resp := model.CursorResponse{}
	if len(items) > size {
		items = items[:size]
		resp.HasMore = true
		resp.NextPageToken = strconv.Itoa(offset + size)
	}

	return items, resp, nil
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
