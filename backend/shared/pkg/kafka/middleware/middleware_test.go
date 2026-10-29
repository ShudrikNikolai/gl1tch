package middleware

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka"
)

var discard = slog.New(slog.NewTextHandler(io.Discard, nil))

var fast = RetryConfig{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond}

type fakeProducer struct {
	mu      sync.Mutex
	sent    []sentMsg
	failing bool
}

type sentMsg struct {
	key, value []byte
	headers    map[string]string
}

func (p *fakeProducer) Send(_ context.Context, key, value []byte, headers ...kafka.Header) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.failing {
		return errors.New("kafka down")
	}
	h := map[string]string{}
	for _, x := range headers {
		h[x.Key] = string(x.Value)
	}
	p.sent = append(p.sent, sentMsg{key, value, h})

	return nil
}

func msg() kafka.Message {
	return kafka.Message{Topic: "posts.events", Partition: 2, Offset: 42, Key: []byte("k"), Value: []byte("v"),
		Headers: map[string][]byte{"trace": []byte("t1")}}
}

func TestRetrySucceedsAfterFailures(t *testing.T) {
	calls := 0
	h := Retry(fast, discard)(func(context.Context, kafka.Message) error {
		calls++
		if calls < 3 {
			return errors.New("flaky")
		}
		return nil
	})
	if err := h(context.Background(), msg()); err != nil || calls != 3 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}

func TestRetryGivesUpAfterMaxAttempts(t *testing.T) {
	calls := 0
	boom := errors.New("boom")
	h := Retry(fast, discard)(func(context.Context, kafka.Message) error { calls++; return boom })
	err := h(context.Background(), msg())
	if !errors.Is(err, boom) || calls != 3 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}

func TestRetrySkipsPermanentErrors(t *testing.T) {
	calls := 0
	h := Retry(fast, discard)(func(context.Context, kafka.Message) error { calls++; return Permanent(errors.New("bad payload")) })
	err := h(context.Background(), msg())
	if err == nil || calls != 1 || !IsPermanent(err) {
		t.Fatalf("err=%v calls=%d permanent=%v", err, calls, IsPermanent(err))
	}
}

func TestRetryStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	h := Retry(RetryConfig{MaxAttempts: 10, BaseDelay: time.Hour, MaxDelay: time.Hour}, discard)(
		func(context.Context, kafka.Message) error { calls++; cancel(); return errors.New("x") })
	done := make(chan error, 1)
	go func() { done <- h(ctx, msg()) }()
	select {
	case err := <-done:
		if err == nil || calls != 1 {
			t.Fatalf("err=%v calls=%d", err, calls)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("retry did not stop on cancel")
	}
}

func TestBackoffBounds(t *testing.T) {
	cfg := RetryConfig{BaseDelay: 100 * time.Millisecond, MaxDelay: 500 * time.Millisecond}
	for attempt := 1; attempt <= 70; attempt++ { // 70: проверка переполнения сдвига
		d := backoff(cfg, attempt)
		if d <= 0 || d > cfg.MaxDelay {
			t.Fatalf("attempt %d: delay %v out of bounds", attempt, d)
		}
	}
}

func TestDLQSwallowsErrorAndKeepsContext(t *testing.T) {
	p := &fakeProducer{}
	h := DLQ(p, discard)(func(context.Context, kafka.Message) error { return errors.New("cannot process") })
	if err := h(context.Background(), msg()); err != nil {
		t.Fatalf("dlq must swallow handled failure, got %v", err)
	}
	if len(p.sent) != 1 {
		t.Fatalf("expected 1 dlq message, got %d", len(p.sent))
	}
	s := p.sent[0]
	if string(s.key) != "k" || string(s.value) != "v" {
		t.Fatalf("original payload lost: %+v", s)
	}
	if s.headers["x-dlq-error"] != "cannot process" || s.headers["x-dlq-source-topic"] != "posts.events" ||
		s.headers["x-dlq-source-partition"] != "2" || s.headers["x-dlq-source-offset"] != "42" || s.headers["trace"] != "t1" {
		t.Fatalf("unexpected headers: %v", s.headers)
	}
}

func TestDLQTruncatesLongError(t *testing.T) {
	p := &fakeProducer{}
	long := errors.New(strings.Repeat("x", 5000))
	h := DLQ(p, discard)(func(context.Context, kafka.Message) error { return long })
	_ = h(context.Background(), msg())
	if got := len(p.sent[0].headers["x-dlq-error"]); got != maxErrorHeader {
		t.Fatalf("error header length %d", got)
	}
}

func TestDLQReturnsErrorWhenSendFails(t *testing.T) {
	p := &fakeProducer{failing: true}
	orig := errors.New("handler failed")
	h := DLQ(p, discard)(func(context.Context, kafka.Message) error { return orig })
	err := h(context.Background(), msg())
	if !errors.Is(err, orig) {
		t.Fatalf("original error must be returned for redelivery, got %v", err)
	}
}

func TestDLQDoesNotCaptureOnShutdown(t *testing.T) {
	p := &fakeProducer{}
	ctx, cancel := context.WithCancel(context.Background())
	h := DLQ(p, discard)(func(context.Context, kafka.Message) error { cancel(); return errors.New("interrupted") })
	if err := h(ctx, msg()); err == nil {
		t.Fatal("error must propagate on shutdown")
	}
	if len(p.sent) != 0 {
		t.Fatal("message must not go to dlq on shutdown")
	}
}

func TestDLQPassesSuccess(t *testing.T) {
	p := &fakeProducer{}
	h := DLQ(p, discard)(func(context.Context, kafka.Message) error { return nil })
	if err := h(context.Background(), msg()); err != nil || len(p.sent) != 0 {
		t.Fatalf("err=%v sent=%d", err, len(p.sent))
	}
}

func TestRecoverTurnsPanicIntoError(t *testing.T) {
	h := Recover(discard)(func(context.Context, kafka.Message) error { panic("oops") })
	err := h(context.Background(), msg())
	if err == nil || !strings.Contains(err.Error(), "oops") {
		t.Fatalf("got %v", err)
	}
}

type fakeStore struct {
	keys    map[string]bool
	delErr  error
	setErr  error
	deleted []string
}

func (s *fakeStore) SetNX(_ context.Context, key string, _ any, _ time.Duration) (bool, error) {
	if s.setErr != nil {
		return false, s.setErr
	}
	if s.keys[key] {
		return false, nil
	}
	s.keys[key] = true

	return true, nil
}

func (s *fakeStore) Del(_ context.Context, key string) error {
	s.deleted = append(s.deleted, key)
	delete(s.keys, key)

	return s.delErr
}

func idFromKey(m kafka.Message) (string, error) {
	if len(m.Key) == 0 {
		return "", errors.New("no event id")
	}

	return string(m.Key), nil
}

func TestDedupeSkipsDuplicates(t *testing.T) {
	store := &fakeStore{keys: map[string]bool{}}
	calls := 0
	h := Dedupe(store, "inbox:", time.Minute, idFromKey, discard)(func(context.Context, kafka.Message) error { calls++; return nil })
	for i := 0; i < 3; i++ {
		if err := h(context.Background(), msg()); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 {
		t.Fatalf("handler must run once, ran %d times", calls)
	}
	if !store.keys["inbox:k"] {
		t.Fatal("event key must stay claimed after success")
	}
}

func TestDedupeReleasesKeyOnFailure(t *testing.T) {
	store := &fakeStore{keys: map[string]bool{}}
	fail := true
	calls := 0
	h := Dedupe(store, "inbox:", time.Minute, idFromKey, discard)(func(context.Context, kafka.Message) error {
		calls++
		if fail {
			return errors.New("temporary")
		}
		return nil
	})
	if err := h(context.Background(), msg()); err == nil {
		t.Fatal("error must propagate")
	}
	if len(store.deleted) != 1 {
		t.Fatalf("key must be released, deleted=%v", store.deleted)
	}
	fail = false
	if err := h(context.Background(), msg()); err != nil || calls != 2 {
		t.Fatalf("retry after release must run the handler: err=%v calls=%d", err, calls)
	}
}

func TestDedupeBadIDIsPermanent(t *testing.T) {
	store := &fakeStore{keys: map[string]bool{}}
	h := Dedupe(store, "inbox:", time.Minute, idFromKey, discard)(func(context.Context, kafka.Message) error { return nil })
	err := h(context.Background(), kafka.Message{})
	if !IsPermanent(err) {
		t.Fatalf("want permanent error, got %v", err)
	}
}

func TestDedupeStoreFailureIsRetryable(t *testing.T) {
	store := &fakeStore{keys: map[string]bool{}, setErr: errors.New("redis down")}
	called := false
	h := Dedupe(store, "inbox:", time.Minute, idFromKey, discard)(func(context.Context, kafka.Message) error { called = true; return nil })
	err := h(context.Background(), msg())
	if err == nil || IsPermanent(err) || called {
		t.Fatalf("err=%v permanent=%v called=%v", err, IsPermanent(err), called)
	}
}

// Полная цепочка в рекомендованном порядке: DLQ -> Retry -> Dedupe -> Recover.
func TestChainPanicEndsInDLQ(t *testing.T) {
	p := &fakeProducer{}
	store := &fakeStore{keys: map[string]bool{}}
	calls := 0
	var h kafka.MessageHandler = func(context.Context, kafka.Message) error { calls++; panic("bug") }
	chain := []kafka.Middleware{
		DLQ(p, discard), Retry(fast, discard), Dedupe(store, "i:", time.Minute, idFromKey, discard), Recover(discard),
	}
	for i := len(chain) - 1; i >= 0; i-- {
		h = chain[i](h)
	}
	if err := h(context.Background(), msg()); err != nil {
		t.Fatalf("chain must end in dlq without error, got %v", err)
	}
	if calls != 3 || len(p.sent) != 1 {
		t.Fatalf("calls=%d dlq=%d", calls, len(p.sent))
	}
	if store.keys["i:k"] {
		t.Fatal("dedupe key must be released after failure")
	}
}
