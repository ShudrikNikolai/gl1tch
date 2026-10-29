package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka"
)

type RetryConfig struct {
	MaxAttempts int           // всего попыток, включая первую (по умолчанию 3)
	BaseDelay   time.Duration // пауза перед второй попыткой (по умолчанию 200ms), дальше удваивается
	MaxDelay    time.Duration // потолок паузы (по умолчанию 5s)
}

func (c RetryConfig) withDefaults() RetryConfig {
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 3
	}
	if c.BaseDelay <= 0 {
		c.BaseDelay = 200 * time.Millisecond
	}
	if c.MaxDelay <= 0 {
		c.MaxDelay = 5 * time.Second
	}

	return c
}

// Retry повторяет обработчик с экспоненциальной паузой и джиттером. Не повторяет ошибки,
// помеченные Permanent, и останавливается при отмене ctx.
func Retry(cfg RetryConfig, log *slog.Logger) kafka.Middleware {
	cfg = cfg.withDefaults()

	return func(next kafka.MessageHandler) kafka.MessageHandler {
		return func(ctx context.Context, msg kafka.Message) error {
			var err error
			for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
				if err = next(ctx, msg); err == nil {
					return nil
				}
				if IsPermanent(err) || ctx.Err() != nil || attempt == cfg.MaxAttempts {
					break
				}

				delay := backoff(cfg, attempt)
				log.WarnContext(ctx, "kafka handler failed, will retry",
					slog.String("topic", msg.Topic), slog.Int64("offset", msg.Offset),
					slog.Int("attempt", attempt), slog.Duration("delay", delay), slog.Any("error", err))

				select {
				case <-time.After(delay):
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			return fmt.Errorf("kafka handler: %w", err)
		}
	}
}

// backoff: base * 2^(attempt-1), не больше MaxDelay, с «полным» джиттером в диапазоне [d/2, d].
func backoff(cfg RetryConfig, attempt int) time.Duration {
	d := cfg.BaseDelay << (attempt - 1)
	if d <= 0 || d > cfg.MaxDelay { // d <= 0: переполнение сдвига
		d = cfg.MaxDelay
	}

	return d/2 + time.Duration(rand.Int64N(int64(d/2)+1))
}
