package consumer

import (
	"context"
	"log/slog"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/notifications/internal/hub"
	sharedkafka "github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka/consumer"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka/middleware"
	eventsv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/events/v1"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Handler struct {
	hub *hub.Hub
	log *slog.Logger
}

func NewHandler(h *hub.Hub, log *slog.Logger) *Handler {
	return &Handler{hub: h, log: log}
}

func (h *Handler) Handle(ctx context.Context, msg sharedkafka.Message) error {
	_, span := otel.Tracer("notifications").Start(ctx, "kafka.post.created")
	defer span.End()

	var event eventsv1.PostCreated
	if err := proto.Unmarshal(msg.Value, &event); err != nil {
		return middleware.Permanent(err)
	}

	data, err := protojson.Marshal(&event)
	if err != nil {
		return middleware.Permanent(err)
	}

	h.hub.Publish(hub.Event{ID: event.GetEventId(), ThreadID: event.GetThreadId(), Data: data})

	return nil
}

// Run переподключается к Kafka с паузой: сервис остаётся доступным для SSE даже без брокера.
func Run(ctx context.Context, brokers []string, topic, group string, handler sharedkafka.MessageHandler, log *slog.Logger) {
	backoff := time.Second
	for ctx.Err() == nil {
		if len(brokers) == 0 {
			log.WarnContext(ctx, "kafka disabled: no brokers")
			return
		}

		cg, err := sharedkafka.NewConsumerGroup(brokers, group, log)
		if err != nil {
			log.WarnContext(ctx, "kafka consumer group failed", slog.Any("error", err))
		} else {
			c := consumer.NewConsumer(cg, []string{topic}, log)
			err = c.Consume(ctx, handler)
			_ = cg.Close()
			if ctx.Err() != nil {
				return
			}
			log.WarnContext(ctx, "kafka consumer stopped", slog.Any("error", err))
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}
