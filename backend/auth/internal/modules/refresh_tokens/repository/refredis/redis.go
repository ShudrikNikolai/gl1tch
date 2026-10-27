package refredis

import (
	"log/slog"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens/repository"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/cache"
)

var _ repository.Repository = (*Repo)(nil)

// Ключи:
//
//	rt:<hash>        JSON model.Session, TTL = время жизни refresh-токена
//	rt:used:<hash>   "<user_id>|<unix_ms>": отметка об использовании, нужна для обнаружения повторного предъявления
//	rt:user:<id>     SET хешей активных токенов пользователя (для отзыва всех сессий)
const (
	prefixSession = "rt:"
	prefixUsed    = "rt:used:"
	prefixUser    = "rt:user:"
)

type Repo struct {
	rdb cache.RedisClient
	log *slog.Logger
}

func New(rdb cache.RedisClient, log *slog.Logger) *Repo {
	return &Repo{rdb: rdb, log: log}
}

func keySession(hash string) string { return prefixSession + hash }
func keyUsed(hash string) string    { return prefixUsed + hash }
func keyUser(userID string) string  { return prefixUser + userID }
