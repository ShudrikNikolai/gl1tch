package consumer

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/IBM/sarama"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka"
)

var discard = slog.New(slog.NewTextHandler(io.Discard, nil))

// Встраивание интерфейсов: тест не ломается, если в sarama появятся новые методы.
type fakeSession struct {
	sarama.ConsumerGroupSession
	ctx    context.Context
	marked []int64
}

func (s *fakeSession) Context() context.Context { return s.ctx }
func (s *fakeSession) MarkMessage(m *sarama.ConsumerMessage, _ string) {
	s.marked = append(s.marked, m.Offset)
}

type fakeClaim struct {
	sarama.ConsumerGroupClaim
	ch chan *sarama.ConsumerMessage
}

func (c *fakeClaim) Messages() <-chan *sarama.ConsumerMessage { return c.ch }

func claimWith(offsets ...int64) *fakeClaim {
	ch := make(chan *sarama.ConsumerMessage, len(offsets))
	for _, o := range offsets {
		ch <- &sarama.ConsumerMessage{
			Topic: "posts.events", Partition: 1, Offset: o, Key: []byte("k"), Value: []byte("v"),
			Headers: []*sarama.RecordHeader{{Key: []byte("h"), Value: []byte("1")}, nil},
		}
	}
	close(ch)

	return &fakeClaim{ch: ch}
}

func TestConsumeClaimMarksProcessedMessages(t *testing.T) {
	var seen []kafka.Message
	gh := NewGroupHandler(func(_ context.Context, m kafka.Message) error { seen = append(seen, m); return nil }, discard)
	sess := &fakeSession{ctx: context.Background()}

	if err := gh.ConsumeClaim(sess, claimWith(10, 11, 12)); err != nil {
		t.Fatal(err)
	}
	if len(sess.marked) != 3 || sess.marked[0] != 10 || sess.marked[2] != 12 {
		t.Fatalf("marked %v", sess.marked)
	}
	if seen[0].Topic != "posts.events" || string(seen[0].Headers["h"]) != "1" || seen[0].Offset != 10 {
		t.Fatalf("message mapped wrongly: %+v", seen[0])
	}
}

// Регрессия: ошибка обработчика не должна приводить к тому, что последующие сообщения
// коммитят offset поверх необработанного.
func TestConsumeClaimStopsOnHandlerErrorAndDoesNotMarkLater(t *testing.T) {
	boom := errors.New("boom")
	gh := NewGroupHandler(func(_ context.Context, m kafka.Message) error {
		if m.Offset == 11 {
			return boom
		}
		return nil
	}, discard)
	sess := &fakeSession{ctx: context.Background()}

	err := gh.ConsumeClaim(sess, claimWith(10, 11, 12))
	if !errors.Is(err, boom) {
		t.Fatalf("handler error must end the claim, got %v", err)
	}
	if len(sess.marked) != 1 || sess.marked[0] != 10 {
		t.Fatalf("only the message before the failure may be marked, got %v", sess.marked)
	}
}

func TestConsumeClaimStopsWhenSessionDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	gh := NewGroupHandler(func(context.Context, kafka.Message) error { return nil }, discard)
	sess := &fakeSession{ctx: ctx}
	claim := &fakeClaim{ch: make(chan *sarama.ConsumerMessage)} // пустой, не закрыт

	done := make(chan error, 1)
	go func() { done <- gh.ConsumeClaim(sess, claim) }()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ConsumeClaim did not stop when the session ended")
	}
}

func TestMiddlewareOrderFirstIsOutermost(t *testing.T) {
	var order []string
	mk := func(name string) kafka.Middleware {
		return func(next kafka.MessageHandler) kafka.MessageHandler {
			return func(ctx context.Context, m kafka.Message) error {
				order = append(order, name+"-in")
				err := next(ctx, m)
				order = append(order, name+"-out")
				return err
			}
		}
	}
	gh := NewGroupHandler(func(context.Context, kafka.Message) error { order = append(order, "handler"); return nil },
		discard, mk("outer"), mk("inner"))
	_ = gh.ConsumeClaim(&fakeSession{ctx: context.Background()}, claimWith(1))

	want := []string{"outer-in", "inner-in", "handler", "inner-out", "outer-out"}
	if len(order) != len(want) {
		t.Fatalf("order %v", order)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order %v, want %v", order, want)
		}
	}
}

// fakeGroup проверяет цикл Consume: ErrClosedConsumerGroup завершает его без ошибки.
type fakeGroup struct {
	sarama.ConsumerGroup
	errs []error
	n    int
}

func (g *fakeGroup) Consume(context.Context, []string, sarama.ConsumerGroupHandler) error {
	err := g.errs[g.n]
	g.n++

	return err
}

func TestConsumeLoop(t *testing.T) {
	handler := func(context.Context, kafka.Message) error { return nil }

	closed := &fakeGroup{errs: []error{nil, sarama.ErrClosedConsumerGroup}}
	if err := NewConsumer(closed, []string{"t"}, discard).Consume(context.Background(), handler); err != nil {
		t.Fatalf("closed group must end quietly, got %v", err)
	}
	if closed.n != 2 {
		t.Fatalf("expected a re-consume after the rebalance, calls=%d", closed.n)
	}

	boom := errors.New("broker error")
	failing := &fakeGroup{errs: []error{boom}}
	if err := NewConsumer(failing, []string{"t"}, discard).Consume(context.Background(), handler); !errors.Is(err, boom) {
		t.Fatalf("got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cancelled := &fakeGroup{errs: []error{nil}}
	if err := NewConsumer(cancelled, []string{"t"}, discard).Consume(ctx, handler); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}
