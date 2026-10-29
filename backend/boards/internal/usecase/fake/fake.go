// Package fake содержит репозитории в памяти для тестов usecase. В продакшен-код не подключается.
package fake

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

// Tx выполняет функцию как есть и считает вызовы. Откат не имитируется.
type Tx struct{ Calls int }

func (t *Tx) InTx(ctx context.Context, fn func(context.Context) error) error {
	t.Calls++

	return fn(ctx)
}

// Users профили по id.
type Users struct{ Items map[string]models.User }

func NewUsers(us ...models.User) *Users {
	f := &Users{Items: map[string]models.User{}}
	for _, u := range us {
		f.Items[u.ID] = u
	}

	return f
}

func (f *Users) Get(_ context.Context, id string) (models.User, error) {
	u, ok := f.Items[id]
	if !ok {
		return models.User{}, apperr.NotFound("USER_NOT_FOUND", "user not found")
	}

	return u, nil
}

// Boards.
type Boards struct{ Items map[string]boards.Board }

func NewBoards() *Boards { return &Boards{Items: map[string]boards.Board{}} }

func (f *Boards) Create(_ context.Context, b boards.Board) (boards.Board, error) {
	for _, x := range f.Items {
		if strings.EqualFold(x.Slug, b.Slug) {
			return boards.Board{}, boards.ErrSlugTaken
		}
	}
	f.Items[b.ID] = b

	return b, nil
}

func (f *Boards) GetByID(_ context.Context, id string) (boards.Board, error) {
	b, ok := f.Items[id]
	if !ok {
		return boards.Board{}, boards.ErrNotFound
	}

	return b, nil
}

func (f *Boards) List(_ context.Context, includeModerator bool) ([]boards.Board, error) {
	var out []boards.Board
	for _, b := range f.Items {
		if b.VisibleTo(includeModerator) {
			out = append(out, b)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SortOrder < out[j].SortOrder })

	return out, nil
}

func (f *Boards) Apply(_ context.Context, id string, d boards.Delta) error {
	b, ok := f.Items[id]
	if !ok {
		return boards.ErrNotFound
	}
	b.ThreadCount += d.Threads
	b.PostCount += d.Posts
	if d.LastPost != nil {
		a, at := d.LastPost.Author, d.LastPost.At
		b.LastPostAuthor, b.LastPostAt = &a, &at
	}
	f.Items[id] = b

	return nil
}

// Threads.
type Threads struct {
	Items map[string]threads.Thread
	seq   map[string]int
}

func NewThreads() *Threads {
	return &Threads{Items: map[string]threads.Thread{}, seq: map[string]int{}}
}

// Seed кладёт готовую тему в хранилище. posts: сколько номеров постов уже выдано.
func (f *Threads) Seed(t threads.Thread, posts int) {
	f.Items[t.ID] = t
	f.seq[t.ID] = posts
}

func (f *Threads) Create(_ context.Context, t threads.Thread) (threads.Thread, error) {
	for _, x := range f.Items {
		if x.BoardID == t.BoardID && x.Slug == t.Slug {
			return threads.Thread{}, threads.ErrSlugTaken
		}
	}
	f.Items[t.ID] = t

	return t, nil
}

func (f *Threads) GetByID(_ context.Context, id string) (threads.Thread, error) {
	t, ok := f.Items[id]
	if !ok || t.DeletedAt != nil {
		return threads.Thread{}, threads.ErrNotFound
	}

	return t, nil
}

// List применяет только фильтры по доске, видимости, категории и статусу, без сортировки и страниц.
func (f *Threads) List(_ context.Context, fl threads.ListFilter, _ domain.PageRequest) ([]threads.Thread, domain.PageInfo, error) {
	var out []threads.Thread
	for _, t := range f.Items {
		switch {
		case t.DeletedAt != nil, t.BoardID != fl.BoardID, !t.VisibleTo(fl.Viewer):
		case fl.Category != "" && t.Category != fl.Category, fl.Status != "" && t.Status != fl.Status:
		default:
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })

	return out, domain.PageInfo{}, nil
}

func (f *Threads) Update(ctx context.Context, id string, p threads.UpdateParams, at time.Time) (threads.Thread, error) {
	t, err := f.GetByID(ctx, id)
	if err != nil {
		return t, err
	}
	if p.Title != nil {
		t.Title = *p.Title
	}
	if p.Category != nil {
		t.Category = *p.Category
	}
	if p.Visibility != nil {
		t.Visibility = *p.Visibility
	}
	t.UpdatedAt = at
	f.Items[id] = t

	return t, nil
}

func (f *Threads) SetStatus(ctx context.Context, id string, s threads.Status, at time.Time) (threads.Thread, error) {
	t, err := f.GetByID(ctx, id)
	if err != nil {
		return t, err
	}
	t.Status, t.UpdatedAt = s, at
	f.Items[id] = t

	return t, nil
}

func (f *Threads) SetPinned(ctx context.Context, id string, pinned bool, at time.Time) (threads.Thread, error) {
	t, err := f.GetByID(ctx, id)
	if err != nil {
		return t, err
	}
	t.Pinned, t.UpdatedAt = pinned, at
	f.Items[id] = t

	return t, nil
}

func (f *Threads) SoftDelete(ctx context.Context, id, _ string, at time.Time) error {
	t, err := f.GetByID(ctx, id)
	if err != nil {
		return err
	}
	t.DeletedAt = &at
	f.Items[id] = t

	return nil
}

func (f *Threads) NextPostNumber(ctx context.Context, id string, author domain.Author, at time.Time) (int, error) {
	t, err := f.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}
	f.seq[id]++
	t.PostCount++
	t.LastPostAuthor, t.LastPostAt = author, at
	f.Items[id] = t

	return f.seq[id], nil
}

func (f *Threads) ReleasePost(ctx context.Context, id string, _ time.Time) error {
	t, err := f.GetByID(ctx, id)
	if err != nil {
		return err
	}
	t.PostCount--
	f.Items[id] = t

	return nil
}

// Posts.
type Posts struct{ Items map[string]posts.Post }

func NewPosts() *Posts { return &Posts{Items: map[string]posts.Post{}} }

func (f *Posts) Create(_ context.Context, p posts.Post) (posts.Post, error) {
	f.Items[p.ID] = p

	return p, nil
}

func (f *Posts) GetByID(_ context.Context, id string) (posts.Post, error) {
	p, ok := f.Items[id]
	if !ok {
		return posts.Post{}, posts.ErrNotFound
	}

	return p, nil
}

func (f *Posts) ListByThread(_ context.Context, threadID string, _ domain.PageRequest) ([]posts.Post, domain.PageInfo, error) {
	var out []posts.Post
	for _, p := range f.Items {
		if p.ThreadID == threadID {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })

	return out, domain.PageInfo{}, nil
}

func (f *Posts) UpdateBody(ctx context.Context, id, body string, at time.Time) (posts.Post, error) {
	p, err := f.GetByID(ctx, id)
	if err != nil || p.Deleted() {
		return posts.Post{}, posts.ErrNotFound
	}
	p.Body, p.EditedAt, p.UpdatedAt = body, &at, at
	f.Items[id] = p

	return p, nil
}

func (f *Posts) SoftDelete(ctx context.Context, id, by string, at time.Time) (posts.Post, bool, error) {
	p, err := f.GetByID(ctx, id)
	if err != nil {
		return posts.Post{}, false, err
	}
	if p.Deleted() {
		return p, false, nil
	}
	p.DeletedAt, p.DeletedBy = &at, by
	f.Items[id] = p

	return p, true, nil
}
