package consumer

import (
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka"
)

// groupHandler адаптер kafka.MessageHandler к sarama.ConsumerGroupHandler.
type groupHandler struct {
	handler kafka.MessageHandler
	log     *slog.Logger
}

// NewGroupHandler применяет цепочку middleware (первый элемент самый внешний) и возвращает обработчик группы.
func NewGroupHandler(handler kafka.MessageHandler, log *slog.Logger, middlewares ...kafka.Middleware) sarama.ConsumerGroupHandler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return &groupHandler{handler: handler, log: log}
}

func (g *groupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (g *groupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (g *groupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				g.log.InfoContext(session.Context(), "kafka message channel closed")
				return nil
			}

			msg := kafka.Message{
				Key:            message.Key,
				Value:          message.Value,
				Topic:          message.Topic,
				Partition:      message.Partition,
				Offset:         message.Offset,
				Timestamp:      message.Timestamp,
				BlockTimestamp: message.BlockTimestamp,
				Headers:        extractHeaders(message.Headers),
			}

			if err := g.handler(session.Context(), msg); err != nil {
				g.log.ErrorContext(session.Context(), "kafka handler failed, ending session for redelivery",
					slog.String("topic", message.Topic), slog.Int("partition", int(message.Partition)),
					slog.Int64("offset", message.Offset), slog.Any("error", err))
				return err
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			g.log.InfoContext(session.Context(), "kafka session context done")
			return nil
		}
	}
}

func extractHeaders(headers []*sarama.RecordHeader) map[string][]byte {
	result := make(map[string][]byte, len(headers))
	for _, h := range headers {
		if h != nil && h.Key != nil {
			result[string(h.Key)] = h.Value
		}
	}

	return result
}
