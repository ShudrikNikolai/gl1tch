package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka"
)

// DedupeStore подходит cache.RedisClient (методы SetNX и Del).
type DedupeStore interface {
	SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error)
	Del(ctx context.Context, key string) error
}

// IDFunc достаёт идентификатор события (event_id) из сообщения, например разбирая Envelope.
type IDFunc func(msg kafka.Message) (string, error)

// Dedupe пропускает уже обработанные события. Идентификатор «занимается» до обработки; если обработчик
// вернул ошибку, занятость снимается, и повтор сможет отработать. Если идентификатор достать не удалось,
// возвращается Permanent-ошибка (сообщение сразу уходит в DLQ). Если хранилище недоступно, возвращается
// обычная ошибка (сообщение не обрабатывается «вслепую»).
func Dedupe(store DedupeStore, prefix string, ttl time.Duration, idOf IDFunc, log *slog.Logger) kafka.Middleware {
	return func(next kafka.MessageHandler) kafka.MessageHandler {
		return func(ctx context.Context, msg kafka.Message) error {
			id, err := idOf(msg)
			if err != nil {
				return Permanent(err)
			}
			key := prefix + id

			claimed, err := store.SetNX(ctx, key, "1", ttl)
			if err != nil {
				return err
			}
			if !claimed {
				log.DebugContext(ctx, "kafka: duplicate event skipped", slog.String("event_id", id))
				return nil
			}

			if err := next(ctx, msg); err != nil {
				// снимаем занятость даже если ctx уже отменён
				relCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
				defer cancel()
				if delErr := store.Del(relCtx, key); delErr != nil {
					log.WarnContext(ctx, "kafka: release dedupe key failed",
						slog.String("event_id", id), slog.Any("error", delErr))
				}

				return err
			}

			return nil
		}
	}
}
