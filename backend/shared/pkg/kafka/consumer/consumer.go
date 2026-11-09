package consumer

import (
	"context"
	"errors"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka"
)

type Middleware = kafka.Middleware

var _ kafka.Consumer = (*Consumer)(nil)

type Consumer struct {
	group       sarama.ConsumerGroup
	topics      []string
	log         *slog.Logger
	middlewares []kafka.Middleware
}

// NewConsumer создаёт consumer группы. middlewares: первый самый внешний
func NewConsumer(group sarama.ConsumerGroup, topics []string, log *slog.Logger, middlewares ...kafka.Middleware) *Consumer {
	if log == nil {
		log = slog.Default()
	}

	return &Consumer{group: group, topics: topics, log: log, middlewares: middlewares}
}

// Consume читает топики до отмены ctx или закрытия группы.
//
// Если обработчик (с учётом middleware) вернул ошибку, сессия завершается и группа перебалансируется:
// сообщение будет доставлено заново с последнего закоммиченного offset. Тихой потери сообщений нет;
// для «хуевых» сообщений используйте middleware.Retry и middleware.DLQ.
func (c *Consumer) Consume(ctx context.Context, handler kafka.MessageHandler) error {
	gh := NewGroupHandler(handler, c.log, c.middlewares...)

	for {
		if err := c.group.Consume(ctx, c.topics, gh); err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) {
				return nil
			}

			c.log.ErrorContext(ctx, "kafka consume error", slog.Any("error", err))
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		c.log.InfoContext(ctx, "kafka consumer group rebalancing")
	}
}
