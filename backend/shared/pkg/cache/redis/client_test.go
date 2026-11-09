package redis

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/cache"
	"github.com/alicebob/miniredis/v2"
)

func newTestClient(t *testing.T) (*Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	c := New(Config{
		Addr: mr.Addr(), MaxIdle: 2, MaxActive: 4,
		DialTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: time.Second, ConnectionTimeout: time.Second,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { _ = c.Close() })

	return c, mr
}

func TestGetNotFound(t *testing.T) {
	c, _ := newTestClient(t)
	_, err := c.Get(context.Background(), "missing")
	if !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestSetGetDelExists(t *testing.T) {
	c, _ := newTestClient(t)
	ctx := context.Background()
	if err := c.Set(ctx, "k", "v"); err != nil {
		t.Fatal(err)
	}
	got, err := c.Get(ctx, "k")
	if err != nil || string(got) != "v" {
		t.Fatalf("get: %q %v", got, err)
	}
	if ok, _ := c.Exists(ctx, "k"); !ok {
		t.Fatal("exists must be true")
	}
	if err := c.Del(ctx, "k"); err != nil {
		t.Fatal(err)
	}
	if ok, _ := c.Exists(ctx, "k"); ok {
		t.Fatal("exists must be false after del")
	}
}

func TestSetWithTTLSubSecond(t *testing.T) {
	c, mr := newTestClient(t)
	ctx := context.Background()
	// раньше int(ttl.Seconds()) давал 0 и Redis отвечал ошибкой
	if err := c.SetWithTTL(ctx, "k", "v", 500*time.Millisecond); err != nil {
		t.Fatalf("sub-second ttl must work: %v", err)
	}
	if mr.TTL("k") <= 0 || mr.TTL("k") > time.Second {
		t.Fatalf("unexpected ttl %v", mr.TTL("k"))
	}
	mr.FastForward(time.Second)
	if _, err := c.Get(ctx, "k"); !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("key must expire, got %v", err)
	}
	if err := c.SetWithTTL(ctx, "k", "v", 0); !errors.Is(err, ErrInvalidTTL) {
		t.Fatalf("zero ttl must be rejected, got %v", err)
	}
}

func TestSetNX(t *testing.T) {
	c, _ := newTestClient(t)
	ctx := context.Background()
	ok, err := c.SetNX(ctx, "lock", "1", time.Minute)
	if err != nil || !ok {
		t.Fatalf("first SetNX: %v %v", ok, err)
	}
	ok, err = c.SetNX(ctx, "lock", "2", time.Minute)
	if err != nil || ok {
		t.Fatalf("second SetNX must report false: %v %v", ok, err)
	}
	got, _ := c.Get(ctx, "lock")
	if string(got) != "1" {
		t.Fatalf("value was overwritten: %q", got)
	}
}

func TestIncrExpire(t *testing.T) {
	c, mr := newTestClient(t)
	ctx := context.Background()
	for want := int64(1); want <= 3; want++ {
		got, err := c.Incr(ctx, "n")
		if err != nil || got != want {
			t.Fatalf("incr: %d %v (want %d)", got, err, want)
		}
	}
	if err := c.Expire(ctx, "n", 1500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if mr.TTL("n") <= 0 {
		t.Fatal("ttl not set")
	}
}

func TestHash(t *testing.T) {
	c, _ := newTestClient(t)
	ctx := context.Background()
	if err := c.HashSet(ctx, "h", map[string]string{"a": "1", "b": "2"}); err != nil {
		t.Fatal(err)
	}
	m, err := c.HGetAll(ctx, "h")
	if err != nil || len(m) != 2 || m["a"] != "1" || m["b"] != "2" {
		t.Fatalf("hgetall: %v %v", m, err)
	}
}

func TestSets(t *testing.T) {
	c, _ := newTestClient(t)
	ctx := context.Background()
	_ = c.SAdd(ctx, "s", "x")
	_ = c.SAdd(ctx, "s", "y")
	if ok, _ := c.SIsMember(ctx, "s", "x"); !ok {
		t.Fatal("x must be a member")
	}
	_ = c.SRem(ctx, "s", "x")
	if ok, _ := c.SIsMember(ctx, "s", "x"); ok {
		t.Fatal("x must be removed")
	}
	members, _ := c.SMembers(ctx, "s")
	if len(members) != 1 || members[0] != "y" {
		t.Fatalf("members: %v", members)
	}
}

func TestCanceledContextFailsFast(t *testing.T) {
	c, _ := newTestClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Ping(ctx); err == nil {
		t.Fatal("canceled context must fail")
	}
}

func recv(t *testing.T, ch <-chan cache.Message, d time.Duration) (cache.Message, bool) {
	t.Helper()
	select {
	case m, ok := <-ch:
		return m, ok
	case <-time.After(d):
		t.Fatal("timeout waiting for message")
		return cache.Message{}, false
	}
}

func TestPubSubRoundTrip(t *testing.T) {
	c, _ := newTestClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := c.Subscribe(ctx, "revoked", "other")
	if err != nil {
		t.Fatal(err)
	}
	// подписка уже подтверждена сервером: сообщение сразу после Subscribe не теряется
	if err := c.Publish(ctx, "revoked", []byte("user-1")); err != nil {
		t.Fatal(err)
	}
	m, _ := recv(t, ch, 2*time.Second)
	if m.Channel != "revoked" || string(m.Payload) != "user-1" {
		t.Fatalf("unexpected message %+v", m)
	}

	cancel()
	for range ch { // канал должен закрыться
	}
}

func TestSubscribeValidation(t *testing.T) {
	c, _ := newTestClient(t)
	if _, err := c.Subscribe(context.Background()); err == nil {
		t.Fatal("no channels must be an error")
	}
	plain := NewClient(c.pool, nil, time.Second) // без подписчика
	if _, err := plain.Subscribe(context.Background(), "x"); err == nil {
		t.Fatal("frontend without subscriber must return an error")
	}
}

func TestPubSubReconnects(t *testing.T) {
	c, mr := newTestClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := c.Subscribe(ctx, "revoked")
	if err != nil {
		t.Fatal(err)
	}

	mr.Close() // обрыв
	time.Sleep(200 * time.Millisecond)
	if err := mr.Restart(); err != nil {
		t.Fatal(err)
	}

	// после переподключения (асинхронно) публикации начинают доходить
	deadline := time.After(8 * time.Second)
	for {
		_ = c.Publish(ctx, "revoked", []byte("after"))
		select {
		case m := <-ch:
			if string(m.Payload) != "after" {
				t.Fatalf("unexpected payload %q", m.Payload)
			}
			return
		case <-time.After(300 * time.Millisecond):
		case <-deadline:
			t.Fatal("subscription did not recover after reconnect")
		}
	}
}
