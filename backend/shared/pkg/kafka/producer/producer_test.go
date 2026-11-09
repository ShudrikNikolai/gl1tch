package producer

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka"
)

var discard = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestSendBuildsMessage(t *testing.T) {
	mp := mocks.NewSyncProducer(t, sarama.NewConfig())
	var got *sarama.ProducerMessage
	mp.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(func(m *sarama.ProducerMessage) error {
		got = m
		return nil
	})

	p := NewProducer(mp, "posts.events", discard)
	err := p.Send(context.Background(), []byte("thread-1"), []byte("payload"),
		kafka.Header{Key: "x-a", Value: []byte("1")}, kafka.Header{Key: "x-b", Value: []byte("2")})
	if err != nil {
		t.Fatal(err)
	}
	if got.Topic != "posts.events" {
		t.Fatalf("topic %q", got.Topic)
	}
	key, _ := got.Key.Encode()
	val, _ := got.Value.Encode()
	if string(key) != "thread-1" || string(val) != "payload" {
		t.Fatalf("key=%q value=%q", key, val)
	}
	if len(got.Headers) != 2 || string(got.Headers[0].Key) != "x-a" || string(got.Headers[1].Value) != "2" {
		t.Fatalf("headers: %+v", got.Headers)
	}
	_ = mp.Close()
}

func TestSendWithoutKey(t *testing.T) {
	mp := mocks.NewSyncProducer(t, sarama.NewConfig())
	var got *sarama.ProducerMessage
	mp.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(func(m *sarama.ProducerMessage) error { got = m; return nil })

	if err := NewProducer(mp, "t", discard).Send(context.Background(), nil, []byte("v")); err != nil {
		t.Fatal(err)
	}
	if got.Key != nil {
		t.Fatal("empty key must be sent as no key")
	}
	_ = mp.Close()
}

func TestSendWrapsError(t *testing.T) {
	mp := mocks.NewSyncProducer(t, sarama.NewConfig())
	mp.ExpectSendMessageAndFail(sarama.ErrOutOfBrokers)

	err := NewProducer(mp, "t", discard).Send(context.Background(), []byte("k"), []byte("v"))
	if !errors.Is(err, sarama.ErrOutOfBrokers) {
		t.Fatalf("want wrapped ErrOutOfBrokers, got %v", err)
	}
	_ = mp.Close()
}

func TestSendRespectsCanceledContext(t *testing.T) {
	mp := mocks.NewSyncProducer(t, sarama.NewConfig()) // ожиданий нет: отправки быть не должно
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := NewProducer(mp, "t", discard).Send(ctx, []byte("k"), []byte("v")); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	_ = mp.Close()
}
