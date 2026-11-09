package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka"
)

const maxErrorHeader = 1024

// DLQ отправляет сообщение в dead-letter топик, если обработчик вернул ошибку, и в этом случае
// считает сообщение обработанным (offset сдвигается). Если записать в DLQ не удалось, возвращается
// исходная ошибка: consumer перезапустит сессию и сообщение придёт снова (потерь нет).
// При отмене ctx (остановка сервиса) сообщение в DLQ не отправляется: оно будет доставлено повторно.
//
// dlq это Producer, привязанный к топику <consumer-group>.dlq.
func DLQ(dlq kafka.Producer, log *slog.Logger) kafka.Middleware {
	return func(next kafka.MessageHandler) kafka.MessageHandler {
		return func(ctx context.Context, msg kafka.Message) error {
			err := next(ctx, msg)
			if err == nil || ctx.Err() != nil {
				return err
			}

			if sendErr := dlq.Send(ctx, msg.Key, msg.Value, dlqHeaders(msg, err)...); sendErr != nil {
				log.ErrorContext(ctx, "kafka: send to dlq failed",
					slog.String("topic", msg.Topic), slog.Int64("offset", msg.Offset), slog.Any("error", sendErr))
				return fmt.Errorf("%w (dlq send failed: %w)", err, sendErr)
			}

			log.WarnContext(ctx, "kafka: message moved to dlq",
				slog.String("topic", msg.Topic), slog.Int("partition", int(msg.Partition)),
				slog.Int64("offset", msg.Offset), slog.Any("error", err))

			return nil
		}
	}
}

func dlqHeaders(msg kafka.Message, err error) []kafka.Header {
	text := err.Error()
	if len(text) > maxErrorHeader {
		text = text[:maxErrorHeader]
	}

	hs := []kafka.Header{
		{Key: "x-dlq-error", Value: []byte(text)},
		{Key: "x-dlq-source-topic", Value: []byte(msg.Topic)},
		{Key: "x-dlq-source-partition", Value: []byte(strconv.Itoa(int(msg.Partition)))},
		{Key: "x-dlq-source-offset", Value: []byte(strconv.FormatInt(msg.Offset, 10))},
		{Key: "x-dlq-at", Value: []byte(time.Now().UTC().Format(time.RFC3339))},
	}

	// исходные заголовки сохраняем (в стабильном порядке)
	keys := make([]string, 0, len(msg.Headers))
	for k := range msg.Headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		hs = append(hs, kafka.Header{Key: k, Value: msg.Headers[k]})
	}

	return hs
}
