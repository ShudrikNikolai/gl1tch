package producer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka"
)

var _ kafka.Producer = (*Producer)(nil)

// Producer пишет в один топик (для DLQ и для события агрегата сервиса)
type Producer struct {
	syncProducer sarama.SyncProducer
	topic        string
	log          *slog.Logger
}

func NewProducer(syncProducer sarama.SyncProducer, topic string, log *slog.Logger) *Producer {
	if log == nil {
		log = slog.Default()
	}

	return &Producer{syncProducer: syncProducer, topic: topic, log: log}
}

// Send отправляет сообщение и ждёт подтверждения
// Пустой key означает «без ключа» (round-robin)
// Содержимое сообщения не логируетс
// Ошибка возвращается вызывающему без доп логирования
func (p *Producer) Send(ctx context.Context, key, value []byte, headers ...kafka.Header) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{Topic: p.topic, Value: sarama.ByteEncoder(value)}
	if len(key) > 0 {
		msg.Key = sarama.ByteEncoder(key)
	}
	for _, h := range headers {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{Key: []byte(h.Key), Value: h.Value})
	}

	partition, offset, err := p.syncProducer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("kafka: send to %s: %w", p.topic, err)
	}

	p.log.DebugContext(ctx, "kafka message sent",
		slog.String("topic", p.topic), slog.Int("partition", int(partition)), slog.Int64("offset", offset),
		slog.Int("key_len", len(key)), slog.Int("value_len", len(value)))

	return nil
}
