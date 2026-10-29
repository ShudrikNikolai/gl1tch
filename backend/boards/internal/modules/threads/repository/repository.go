package repository

import (
	"context"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
)

type Repository interface {
	// Create возвращает threads.ErrSlugTaken, если slug занят в этой доске.
	Create(ctx context.Context, t threads.Thread) (threads.Thread, error)
	// GetByID не отдаёт удалённые темы. threads.ErrNotFound, если темы нет или id не uuid.
	GetByID(ctx context.Context, id string) (threads.Thread, error)
	// List: курсорная выдача, закреплённые темы первыми. Токен привязан к порядку из фильтра.
	List(ctx context.Context, f threads.ListFilter, page domain.PageRequest) ([]threads.Thread, domain.PageInfo, error)

	Update(ctx context.Context, id string, p threads.UpdateParams, at time.Time) (threads.Thread, error)
	SetStatus(ctx context.Context, id string, s threads.Status, at time.Time) (threads.Thread, error)
	SetPinned(ctx context.Context, id string, pinned bool, at time.Time) (threads.Thread, error)
	// SoftDelete: threads.ErrNotFound, если темы нет или она уже удалена.
	SoftDelete(ctx context.Context, id, by string, at time.Time) error

	// NextPostNumber выдаёт следующий номер поста и сразу обновляет счётчик и «последний пост» темы.
	// Вызывайте в одной транзакции со вставкой поста: строка темы блокируется до коммита, номера не повторяются.
	NextPostNumber(ctx context.Context, id string, author domain.Author, at time.Time) (int, error)
	// ReleasePost уменьшает счётчик постов при удалении поста. Номера не переиспользуются.
	ReleasePost(ctx context.Context, id string, at time.Time) error
}
