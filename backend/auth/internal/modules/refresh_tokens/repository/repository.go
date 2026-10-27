package repository

import (
	"context"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens"
)

// Repository все методы принимают хеш токена, сырое значение до репозитория не доходит.
type Repository interface {
	Save(ctx context.Context, hash string, s refresh_tokens.Session, ttl time.Duration) error
	// Get возвращает refresh_tokens.ErrSessionNotFound, если токена нет.
	Get(ctx context.Context, hash string) (refresh_tokens.Session, error)
	// Delete удаляет токен. userID нужен, чтобы убрать его из списка сессий пользователя.
	Delete(ctx context.Context, hash, userID string) error
	DeleteAllByUser(ctx context.Context, userID string) error

	// MarkUsed атомарно помечает токен использованным (ротация). false: его уже использовали раньше.
	MarkUsed(ctx context.Context, hash, userID string, at time.Time, ttl time.Duration) (bool, error)
	// Used возвращает, кто и когда использовал токен. found=false: отметки нет.
	Used(ctx context.Context, hash string) (userID string, at time.Time, found bool, err error)
}
