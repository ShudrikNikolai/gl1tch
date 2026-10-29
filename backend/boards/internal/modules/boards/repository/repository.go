package repository

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
)

type Repository interface {
	// Create возвращает boards.ErrSlugTaken, если slug занят (без учёта регистра).
	Create(ctx context.Context, b boards.Board) (boards.Board, error)
	// GetByID возвращает boards.ErrNotFound, если доски нет или id не uuid.
	GetByID(ctx context.Context, id string) (boards.Board, error)
	// List отдаёт доски по sort_order. includeModerator=false скрывает доски только для модераторов.
	List(ctx context.Context, includeModerator bool) ([]boards.Board, error)
	// Apply меняет счётчики и превью последнего поста. Вызывайте в одной транзакции с изменением темы или поста.
	Apply(ctx context.Context, boardID string, d boards.Delta) error
}
