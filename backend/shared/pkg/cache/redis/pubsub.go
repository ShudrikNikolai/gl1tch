package redis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/cache"
	redigo "github.com/gomodule/redigo/redis"
)

const (
	subPingEvery   = 30 * time.Second
	subReadTimeout = subPingEvery + 10*time.Second
	subOpenTimeout = 5 * time.Second
	subBuffer      = 64
	subMinBackoff  = 100 * time.Millisecond
	subMaxBackoff  = 5 * time.Second
)

// Subscribe подписывается на каналы. Подписка переподключается сама (с экспоненциальной паузой),
// результирующий канал закрывается при отмене ctx. К моменту возврата подписка подтверждена сервером,
// поэтому опубликованное после возврата сообщение не потеряется.
func (c *Client) Subscribe(ctx context.Context, channels ...string) (<-chan cache.Message, error) {
	if c.subDial == nil {
		return nil, errors.New("redis: subscriber is not configured (create the frontend with New or WithSubscriber)")
	}
	if len(channels) == 0 {
		return nil, errors.New("redis: no channels to subscribe")
	}

	psc, pending, err := c.openSub(ctx, channels)
	if err != nil {
		return nil, err
	}
	out := make(chan cache.Message, subBuffer)
	go c.subLoop(ctx, psc, pending, channels, out)

	return out, nil
}

// openSub открывает соединение, подписывается и ждёт подтверждений по всем каналам.
// Сообщения, пришедшие раньше последнего подтверждения, возвращаются в pending.
func (c *Client) openSub(ctx context.Context, channels []string) (*redigo.PubSubConn, []redigo.Message, error) {
	openCtx, cancel := context.WithTimeout(ctx, subOpenTimeout)
	defer cancel()

	conn, err := c.subDial(openCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("redis: dial subscriber: %w", err)
	}
	psc := &redigo.PubSubConn{Conn: conn}

	args := make([]any, len(channels))
	for i, ch := range channels {
		args[i] = ch
	}
	if err := psc.Subscribe(args...); err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("redis: subscribe: %w", err)
	}

	var pending []redigo.Message
	for confirmed := 0; confirmed < len(channels); {
		switch v := psc.ReceiveContext(openCtx).(type) {
		case redigo.Subscription:
			if v.Kind == "subscribe" {
				confirmed++
			}
		case redigo.Message:
			pending = append(pending, v)
		case error:
			_ = conn.Close()
			return nil, nil, fmt.Errorf("redis: wait subscription: %w", v)
		}
	}

	return psc, pending, nil
}

func (c *Client) subLoop(ctx context.Context, psc *redigo.PubSubConn, pending []redigo.Message, channels []string, out chan<- cache.Message) {
	defer close(out)

	backoff := subMinBackoff
	for {
		err := c.pump(ctx, psc, pending, out)
		_ = psc.Close()
		if ctx.Err() != nil {
			return
		}
		c.log.WarnContext(ctx, "redis: pubsub disconnected, reconnecting",
			slog.Any("error", err), slog.Any("channels", channels))

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			var oerr error
			psc, pending, oerr = c.openSub(ctx, channels)
			if oerr == nil {
				backoff = subMinBackoff
				break
			}
			c.log.WarnContext(ctx, "redis: pubsub reconnect failed", slog.Any("error", oerr))
			backoff = min(backoff*2, subMaxBackoff)
		}
	}
}

// pump читает сообщения до ошибки соединения или отмены ctx.
func (c *Client) pump(ctx context.Context, psc *redigo.PubSubConn, pending []redigo.Message, out chan<- cache.Message) error {
	stop := make(chan struct{})
	defer close(stop)
	go keepalive(psc, stop)

	deliver := func(m redigo.Message) error {
		select {
		case out <- cache.Message{Channel: m.Channel, Payload: m.Data}:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	for _, m := range pending {
		if err := deliver(m); err != nil {
			return err
		}
	}
	for {
		switch v := psc.ReceiveContext(ctx).(type) {
		case redigo.Message:
			if err := deliver(v); err != nil {
				return err
			}
		case error:
			return v
		}
		// Subscription и Pong не интересны потребителю
	}
}

// keepalive шлёт PING, чтобы обрыв «в тишине» обнаруживался по read timeout, а не висел вечно.
func keepalive(psc *redigo.PubSubConn, stop <-chan struct{}) {
	t := time.NewTicker(subPingEvery)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			if err := psc.Ping(""); err != nil {
				_ = psc.Close()
				return
			}
		}
	}
}
