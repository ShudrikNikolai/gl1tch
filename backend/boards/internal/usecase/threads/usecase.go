package threads

import (
	"context"
	"log/slog"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/access"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

// Интерфейсы объявлены здесь, у потребителя. Реализации: pg-репозитории модулей и pg.Client.

type Tx interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type BoardRepo interface {
	GetByID(ctx context.Context, id string) (boards.Board, error)
	Apply(ctx context.Context, boardID string, d boards.Delta) error
}

type ThreadRepo interface {
	Create(ctx context.Context, t threads.Thread) (threads.Thread, error)
	GetByID(ctx context.Context, id string) (threads.Thread, error)
	List(ctx context.Context, f threads.ListFilter, page domain.PageRequest) ([]threads.Thread, domain.PageInfo, error)
	Update(ctx context.Context, id string, p threads.UpdateParams, at time.Time) (threads.Thread, error)
	SetStatus(ctx context.Context, id string, s threads.Status, at time.Time) (threads.Thread, error)
	SetPinned(ctx context.Context, id string, pinned bool, at time.Time) (threads.Thread, error)
	SoftDelete(ctx context.Context, id, by string, at time.Time) error
	NextPostNumber(ctx context.Context, id string, author domain.Author, at time.Time) (int, error)
}

type PostRepo interface {
	Create(ctx context.Context, p posts.Post) (posts.Post, error)
}

type Deps struct {
	Tx      Tx
	Boards  BoardRepo
	Threads ThreadRepo
	Posts   PostRepo
	Users   access.Users
	Log     *slog.Logger
}

type UseCase struct {
	tx      Tx
	boards  BoardRepo
	threads ThreadRepo
	posts   PostRepo
	users   access.Users
	log     *slog.Logger
	now     func() time.Time
}

func New(d Deps) *UseCase {
	return &UseCase{
		tx:      d.Tx,
		boards:  d.Boards,
		threads: d.Threads,
		posts:   d.Posts,
		users:   d.Users,
		log:     d.Log,
		now:     time.Now,
	}
}

// checkCategory проверяет категорию: FAQ могут создавать и менять только модераторы.
func checkCategory(actor models.Actor, c threads.Category) error {
	if !c.Valid() {
		return apperr.InvalidArgument("INVALID_CATEGORY", "unknown thread category")
	}
	// FAQ ведут модераторы
	if c == threads.CategoryFAQ && !actor.IsModerator() {
		return access.ErrForbidden
	}

	return nil
}

func checkVisibility(v domain.Visibility) error {
	if !v.Valid() {
		return apperr.InvalidArgument("INVALID_VISIBILITY", "unknown visibility")
	}

	return nil
}
