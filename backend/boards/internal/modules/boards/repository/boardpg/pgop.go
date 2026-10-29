package boardpg

import (
	"context"
	"errors"
	"fmt"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/storage/pg"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards/repository/converter"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards/repository/model"
	"github.com/jackc/pgx/v5"
)

type scanner interface {
	Scan(dest ...any) error
}

func scan(s scanner) (boards.Board, error) {
	var m model.Board

	err := s.Scan(
		&m.ID, &m.Slug, &m.Title, &m.Description, &m.Visibility, &m.SortOrder,
		&m.ThreadCount, &m.PostCount, &m.LastPostAuthorID, &m.LastPostAuthorName, &m.LastPostAt,
		&m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return boards.Board{}, err
	}

	return converter.BoardModelToSrv(m), nil
}

func (r *Repo) Create(ctx context.Context, b boards.Board) (boards.Board, error) {
	m := converter.BoardSrvToModel(b)

	created, err := scan(r.db.DB(ctx).QueryRow(ctx, qCreate,
		m.ID, m.Slug, m.Title, m.Description, m.Visibility, m.SortOrder, m.CreatedBy))
	if err != nil {
		if c, ok := pg.UniqueViolation(err); ok && c == slugConstraint {
			return boards.Board{}, boards.ErrSlugTaken
		}

		return boards.Board{}, fmt.Errorf("boards: create: %w", err)
	}

	return created, nil
}

func (r *Repo) GetByID(ctx context.Context, id string) (boards.Board, error) {
	b, err := scan(r.db.DB(ctx).QueryRow(ctx, qGet, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || pg.InvalidInput(err) {
			return boards.Board{}, boards.ErrNotFound
		}

		return boards.Board{}, fmt.Errorf("boards: get: %w", err)
	}

	return b, nil
}

func (r *Repo) List(ctx context.Context, includeModerator bool) ([]boards.Board, error) {
	rows, err := r.db.DB(ctx).Query(ctx, qList, includeModerator)
	if err != nil {
		return nil, fmt.Errorf("boards: list: %w", err)
	}
	defer rows.Close()

	var out []boards.Board
	for rows.Next() {
		b, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("boards: list scan: %w", err)
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("boards: list: %w", err)
	}

	return out, nil
}

func (r *Repo) Apply(ctx context.Context, boardID string, d boards.Delta) error {
	var (
		authorID, authorName *string
		at                   any // nil -> NULL
	)
	if d.LastPost != nil {
		authorID, authorName, at = &d.LastPost.Author.UserID, &d.LastPost.Author.Username, d.LastPost.At
	}

	tag, err := r.db.DB(ctx).Exec(ctx, qApply, boardID, d.Threads, d.Posts, authorID, authorName, at)
	if err != nil {
		if pg.InvalidInput(err) {
			return boards.ErrNotFound
		}

		return fmt.Errorf("boards: apply: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return boards.ErrNotFound
	}

	return nil
}
