package repository

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/credentials"
)

type Repository interface {
	// Create возвращает credentials.ErrUsernameTaken, если логин занят (без учёта регистра).
	Create(ctx context.Context, c credentials.Credential) error
	// GetByUsername возвращает credentials.ErrNotFound, если логина нет.
	GetByUsername(ctx context.Context, username string) (credentials.Credential, error)
	UpdateHash(ctx context.Context, userID, passwordHash string) error
	Delete(ctx context.Context, userID string) error
}
