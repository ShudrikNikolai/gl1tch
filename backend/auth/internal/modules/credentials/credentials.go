package credentials

import (
	"errors"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

type Credential struct {
	UserID       string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

var (
	// ErrNotFound внутренняя ошибка репозитория: наружу её не отдаём (см. ErrInvalidCredentials).
	ErrNotFound = errors.New("credentials: not found")

	ErrUsernameTaken      = apperr.AlreadyExists("USERNAME_TAKEN", "username is already taken")
	ErrInvalidCredentials = apperr.Unauthenticated("INVALID_CREDENTIALS", "invalid username or password")
)
