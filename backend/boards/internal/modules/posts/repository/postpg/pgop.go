package postpg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/cursor"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/storage/pg"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts/repository/converter"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts/repository/model"
	"github.com/jackc/pgx/v5"
)

type scanner interface {
	Scan(dest ...any) error
}

func scan(s scanner) (posts.Post, error) {
	var m model.Post

	err := s.Scan(
		&m.ID, &m.ThreadID, &m.Number, &m.ParentID, &m.AuthorID, &m.AuthorName,
		&m.Body, &m.EditedAt, &m.DeletedAt, &m.DeletedBy, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return posts.Post{}, err
	}

	return converter.PostModelToSrv(m), nil
}

func (r *Repo) Create(ctx context.Context, p posts.Post) (posts.Post, error) {
	m := converter.PostSrvToModel(p)

	created, err := scan(r.db.DB(ctx).QueryRow(ctx, qCreate,
		m.ID, m.ThreadID, m.Number, m.ParentID, m.AuthorID, m.AuthorName, m.Body, m.CreatedAt))
	if err != nil {
		return posts.Post{}, fmt.Errorf("posts: create: %w", err)
	}

	return created, nil
}

func (r *Repo) GetByID(ctx context.Context, id string) (posts.Post, error) {
	p, err := scan(r.db.DB(ctx).QueryRow(ctx, qGet, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || pg.InvalidInput(err) {
			return posts.Post{}, posts.ErrNotFound
		}

		return posts.Post{}, fmt.Errorf("posts: get: %w", err)
	}

	return p, nil
}

func (r *Repo) ListByThread(ctx context.Context, threadID string, page domain.PageRequest) ([]posts.Post, domain.PageInfo, error) {
	var cur cursorData
	if page.Token != "" {
		if err := cursor.Decode(page.Token, cursorKind, &cur); err != nil {
			return nil, domain.PageInfo{}, err
		}
	}

	size := page.EffectiveSize()

	rows, err := r.db.DB(ctx).Query(ctx, qList, threadID, cur.Number, size+1) // на одну строку больше: так узнаём, есть ли следующая страница
	if err != nil {
		if pg.InvalidInput(err) {
			return nil, domain.PageInfo{}, nil
		}

		return nil, domain.PageInfo{}, fmt.Errorf("posts: list: %w", err)
	}
	defer rows.Close()

	items := make([]posts.Post, 0, size+1)
	for rows.Next() {
		p, err := scan(rows)
		if err != nil {
			return nil, domain.PageInfo{}, fmt.Errorf("posts: list scan: %w", err)
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.PageInfo{}, fmt.Errorf("posts: list: %w", err)
	}

	var info domain.PageInfo
	if len(items) > size {
		items = items[:size]
		info.HasMore = true

		tok, err := cursor.Encode(cursorKind, cursorData{Number: items[size-1].Number})
		if err != nil {
			return nil, domain.PageInfo{}, fmt.Errorf("posts: list cursor: %w", err)
		}
		info.NextToken = tok
	}

	return items, info, nil
}

func (r *Repo) UpdateBody(ctx context.Context, id, body string, at time.Time) (posts.Post, error) {
	p, err := scan(r.db.DB(ctx).QueryRow(ctx, qUpdateBody, id, body, at))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || pg.InvalidInput(err) {
			return posts.Post{}, posts.ErrNotFound
		}

		return posts.Post{}, fmt.Errorf("posts: update body: %w", err)
	}

	return p, nil
}

func (r *Repo) HasUserPosted(ctx context.Context, threadID, userID string) (bool, error) {
	var posted bool
	if err := r.db.DB(ctx).QueryRow(ctx, qHasUserPosted, threadID, userID).Scan(&posted); err != nil {
		if pg.InvalidInput(err) {
			return false, nil
		}

		return false, fmt.Errorf("posts: has user posted: %w", err)
	}

	return posted, nil
}

func (r *Repo) SoftDelete(ctx context.Context, id, by string, at time.Time) (posts.Post, bool, error) {
	p, err := scan(r.db.DB(ctx).QueryRow(ctx, qSoftDelete, id, at, by))
	if err == nil {
		return p, true, nil
	}
	if pg.InvalidInput(err) {
		return posts.Post{}, false, posts.ErrNotFound
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return posts.Post{}, false, fmt.Errorf("posts: soft delete: %w", err)
	}

	// Обновлять было нечего: поста нет или он уже удалён
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return posts.Post{}, false, err
	}

	return existing, false, nil
}
