package kafka

import (
	"log/slog"
	"time"

	"github.com/IBM/sarama"
)

// NewSaramaConfig общий конфиг для producer/consumer проекта.
func NewSaramaConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_1_0_0
	cfg.ClientID = "gl1tch"

	cfg.Producer.RequiredAcks = sarama.WaitForLocal
	cfg.Producer.Return.Successes = true
	cfg.Producer.Retry.Max = 3
	cfg.Producer.Timeout = 5 * time.Second

	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	cfg.Consumer.Group.Session.Timeout = 10 * time.Second
	cfg.Consumer.Group.Heartbeat.Interval = 3 * time.Second

	return cfg
}

func NewSyncProducer(brokers []string, log *slog.Logger) (sarama.SyncProducer, error) {
	if log != nil {
		log.Debug("connect kafka producer", slog.Any("brokers", brokers))
	}

	return sarama.NewSyncProducer(brokers, NewSaramaConfig())
}

func NewConsumerGroup(brokers []string, group string, log *slog.Logger) (sarama.ConsumerGroup, error) {
	if log != nil {
		log.Debug("connect kafka consumer", slog.Any("brokers", brokers), slog.String("group", group))
	}

	return sarama.NewConsumerGroup(brokers, group, NewSaramaConfig())
}
