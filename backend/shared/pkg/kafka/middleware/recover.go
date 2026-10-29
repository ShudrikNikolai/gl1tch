package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka"
)

// Recover превращает панику обработчика в ошибку (со стеком в логе), чтобы её обработали Retry и DLQ.
func Recover(log *slog.Logger) kafka.Middleware {
	return func(next kafka.MessageHandler) kafka.MessageHandler {
		return func(ctx context.Context, msg kafka.Message) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.ErrorContext(ctx, "kafka handler panic",
						slog.Any("panic", r), slog.String("topic", msg.Topic),
						slog.Int64("offset", msg.Offset), slog.String("stack", string(debug.Stack())))
					err = fmt.Errorf("handler panic: %v", r)
				}
			}()

			return next(ctx, msg)
		}
	}
}
