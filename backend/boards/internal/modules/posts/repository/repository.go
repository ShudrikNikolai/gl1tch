package repository

import (
	"context"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
)

type Repository interface {
	// Create сохраняет пост с уже выданным номером (см. threads.Repository.NextPostNumber, одна транзакция).
	Create(ctx context.Context, p posts.Post) (posts.Post, error)
	// GetByID отдаёт и удалённые посты (читатели видят заглушку). posts.ErrNotFound, если поста нет или id не uuid.
	GetByID(ctx context.Context, id string) (posts.Post, error)
	// ListByThread отдаёт посты по возрастанию номера, включая удалённые: ответы под ними должны остаться на месте.
	ListByThread(ctx context.Context, threadID string, page domain.PageRequest) ([]posts.Post, domain.PageInfo, error)
	// UpdateBody меняет текст не удалённого поста и ставит edited_at. posts.ErrNotFound, если поста нет или он удалён.
	UpdateBody(ctx context.Context, id, body string, at time.Time) (posts.Post, error)
	// SoftDelete помечает пост удалённым. deleted=false: пост уже был удалён (повтор безопасен, счётчики менять не надо).
	SoftDelete(ctx context.Context, id, by string, at time.Time) (p posts.Post, deleted bool, err error)
}
