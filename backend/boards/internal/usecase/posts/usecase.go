package posts

import (
	"context"
	"log/slog"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/access"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

var (
	ErrInvalidParent = apperr.InvalidArgument("INVALID_PARENT_POST", "parent post must be an existing, not deleted top-level post of the same thread")
	ErrFirstPost     = apperr.InvalidArgument("FIRST_POST", "the first post cannot be deleted separately, delete the thread instead")
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
	GetByID(ctx context.Context, id string) (threads.Thread, error)
	NextPostNumber(ctx context.Context, id string, author domain.Author, at time.Time) (int, error)
	ReleasePost(ctx context.Context, id string, at time.Time) error
}

type PostRepo interface {
	Create(ctx context.Context, p posts.Post) (posts.Post, error)
	GetByID(ctx context.Context, id string) (posts.Post, error)
	ListByThread(ctx context.Context, threadID string, page domain.PageRequest) ([]posts.Post, domain.PageInfo, error)
	UpdateBody(ctx context.Context, id, body string, at time.Time) (posts.Post, error)
	SoftDelete(ctx context.Context, id, by string, at time.Time) (posts.Post, bool, error)
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
