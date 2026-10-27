package refresh_tokens

import (
	"errors"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

// Token выданный токен. Value известно только в момент выдачи.
type Token struct {
	Value     string
	ExpiresAt time.Time
}

// Session то, что хранится по хешу токена.
type Session struct {
	UserID   string
	IssuedAt time.Time
}

var (
	// ErrSessionNotFound внутренняя ошибка репозитория: наружу не отдаётся.
	ErrSessionNotFound = errors.New("refresh_tokens: session not found")

	ErrInvalidToken = apperr.Unauthenticated("INVALID_REFRESH_TOKEN", "invalid refresh token")
	ErrTokenReused  = apperr.Unauthenticated("REFRESH_TOKEN_REUSED", "refresh token reused")
)
