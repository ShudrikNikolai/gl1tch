package threadpg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/cursor"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/storage/pg"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads/repository/converter"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads/repository/model"
	"github.com/jackc/pgx/v5"
)

type scanner interface {
	Scan(dest ...any) error
}

func scan(s scanner) (threads.Thread, error) {
	var m model.Thread

	err := s.Scan(
		&m.ID, &m.BoardID, &m.Title, &m.Slug, &m.AuthorID, &m.AuthorName,
		&m.Status, &m.Category, &m.Visibility, &m.Pinned, &m.PostCount,
		&m.LastPostAuthorID, &m.LastPostAuthorName, &m.LastPostAt,
		&m.DeletedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return threads.Thread{}, err
	}

	return converter.ThreadModelToSrv(m), nil
}

// one выполняет запрос с RETURNING и превращает «ничего не найдено» в threads.ErrNotFound.
func (r *Repo) one(ctx context.Context, op, sql string, args ...any) (threads.Thread, error) {
	t, err := scan(r.db.DB(ctx).QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || pg.InvalidInput(err) {
			return threads.Thread{}, threads.ErrNotFound
		}

		return threads.Thread{}, fmt.Errorf("threads: %s: %w", op, err)
	}

	return t, nil
}

func (r *Repo) Create(ctx context.Context, t threads.Thread) (threads.Thread, error) {
	m := converter.ThreadSrvToModel(t)

	created, err := scan(r.db.DB(ctx).QueryRow(ctx, qCreate,
		m.ID, m.BoardID, m.Title, m.Slug, m.AuthorID, m.AuthorName,
		m.Status, m.Category, m.Visibility, m.Pinned, m.CreatedAt))
	if err != nil {
		if c, ok := pg.UniqueViolation(err); ok && c == slugConstraint {
			return threads.Thread{}, threads.ErrSlugTaken
		}

		return threads.Thread{}, fmt.Errorf("threads: create: %w", err)
	}

	return created, nil
}

func (r *Repo) GetByID(ctx context.Context, id string) (threads.Thread, error) {
	return r.one(ctx, "get", qGet, id)
}

func (r *Repo) List(ctx context.Context, f threads.ListFilter, page domain.PageRequest) ([]threads.Thread, domain.PageInfo, error) {
	if f.Order == "" {
		f.Order = threads.OrderLastPost
	}

	var cur *cursorData
	if page.Token != "" {
		cur = &cursorData{}
		if err := cursor.Decode(page.Token, cursorKind, cur); err != nil {
			return nil, domain.PageInfo{}, err
		}
		if cur.Order != f.Order {
			return nil, domain.PageInfo{}, cursor.ErrInvalid
		}
	}

	size := page.EffectiveSize()
	q, args := buildList(f, size+1, cur) // на одну строку больше: так узнаём, есть ли следующая страница

	rows, err := r.db.DB(ctx).Query(ctx, q, args...)
	if err != nil {
		if pg.InvalidInput(err) { // board_id или id автора не uuid: таких записей нет
			return nil, domain.PageInfo{}, nil
		}

		return nil, domain.PageInfo{}, fmt.Errorf("threads: list: %w", err)
	}
	defer rows.Close()

	items := make([]threads.Thread, 0, size+1)
	for rows.Next() {
		t, err := scan(rows)
		if err != nil {
			return nil, domain.PageInfo{}, fmt.Errorf("threads: list scan: %w", err)
		}
		items = append(items, t)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.PageInfo{}, fmt.Errorf("threads: list: %w", err)
	}

	var info domain.PageInfo
	if len(items) > size {
		items = items[:size]
		info.HasMore = true

		tok, err := cursor.Encode(cursorKind, cursorFrom(f.Order, items[size-1]))
		if err != nil {
			return nil, domain.PageInfo{}, fmt.Errorf("threads: list cursor: %w", err)
		}
		info.NextToken = tok
	}

	return items, info, nil
}

func (r *Repo) Update(ctx context.Context, id string, p threads.UpdateParams, at time.Time) (threads.Thread, error) {
	return r.one(ctx, "update", qUpdate, id, p.Title, strPtr(p.Category), strPtr(p.Visibility), at)
}

func (r *Repo) SetStatus(ctx context.Context, id string, s threads.Status, at time.Time) (threads.Thread, error) {
	return r.one(ctx, "set status", qSetStatus, id, string(s), at)
}

func (r *Repo) SetPinned(ctx context.Context, id string, pinned bool, at time.Time) (threads.Thread, error) {
	return r.one(ctx, "set pinned", qSetPinned, id, pinned, at)
}

func (r *Repo) SoftDelete(ctx context.Context, id, by string, at time.Time) error {
	tag, err := r.db.DB(ctx).Exec(ctx, qSoftDelete, id, at, by)
	if err != nil {
		if pg.InvalidInput(err) {
			return threads.ErrNotFound
		}

		return fmt.Errorf("threads: soft delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return threads.ErrNotFound
	}

	return nil
}

func (r *Repo) NextPostNumber(ctx context.Context, id string, author domain.Author, at time.Time) (int, error) {
	var number int

	err := r.db.DB(ctx).QueryRow(ctx, qNextPostNumber, id, author.UserID, author.Username, at).Scan(&number)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || pg.InvalidInput(err) {
			return 0, threads.ErrNotFound
		}

		return 0, fmt.Errorf("threads: next post number: %w", err)
	}

	return number, nil
}

func (r *Repo) ReleasePost(ctx context.Context, id string, at time.Time) error {
	tag, err := r.db.DB(ctx).Exec(ctx, qReleasePost, id, at)
	if err != nil {
		return fmt.Errorf("threads: release post: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return threads.ErrNotFound
	}

	return nil
}

// strPtr превращает указатель на строковый тип в *string: nil остаётся nil и уходит в запрос как NULL.
func strPtr[T ~string](p *T) *string {
	if p == nil {
		return nil
	}
	s := string(*p)

	return &s
}
