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

var _ cache.RedisClient = (*Client)(nil)

// ErrInvalidTTL: время жизни должно быть положительным.
var ErrInvalidTTL = errors.New("redis: ttl must be positive")

type Client struct {
	pool              *redigo.Pool
	log               *slog.Logger
	connectionTimeout time.Duration
	subDial           func(ctx context.Context) (redigo.Conn, error)
}

type Option func(*Client)

// WithSubscriber задаёт способ открыть отдельное соединение для Subscribe.
// Оно не берётся из пула: подписка держит его всё время, и у него свой read timeout.
func WithSubscriber(dial func(ctx context.Context) (redigo.Conn, error)) Option {
	return func(c *Client) { c.subDial = dial }
}

// NewClient оборачивает готовый пул. connectionTimeout ограничивает ожидание свободного соединения.
// Для Subscribe нужен WithSubscriber; проще создавать клиент через New(cfg, log).
func NewClient(pool *redigo.Pool, log *slog.Logger, connectionTimeout time.Duration, opts ...Option) *Client {
	if log == nil {
		log = slog.Default()
	}
	c := &Client{pool: pool, log: log, connectionTimeout: connectionTimeout}
	for _, o := range opts {
		o(c)
	}

	return c
}

// Close закрывает пул.
func (c *Client) Close() error { return c.pool.Close() }

func (c *Client) getConn(ctx context.Context) (redigo.Conn, error) {
	if c.connectionTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.connectionTimeout)
		defer cancel()
	}
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("redis: get connection: %w", err)
	}

	return conn, nil
}

// do выполняет одну команду с учётом ctx (отмена и deadline прерывают и ожидание ответа).
// Ошибка возвращается наверх с контекстом, без логирования: логирует тот, кто её обрабатывает.
func (c *Client) do(ctx context.Context, cmd string, args ...any) (any, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			c.log.WarnContext(ctx, "redis: close connection", slog.Any("error", cerr))
		}
	}()

	reply, err := redigo.DoContext(conn, ctx, cmd, args...)
	if err != nil {
		return nil, fmt.Errorf("redis %s: %w", cmd, err)
	}

	return reply, nil
}

// ttlMillis переводит TTL в миллисекунды (не меньше 1 мс). Секунды для TTL < 1с давали бы 0 и ошибку Redis.
func ttlMillis(ttl time.Duration) (int64, error) {
	if ttl <= 0 {
		return 0, ErrInvalidTTL
	}

	return max(ttl.Milliseconds(), 1), nil
}

func (c *Client) Set(ctx context.Context, key string, value any) error {
	_, err := c.do(ctx, "SET", key, value)
	return err
}

func (c *Client) SetWithTTL(ctx context.Context, key string, value any, ttl time.Duration) error {
	ms, err := ttlMillis(ttl)
	if err != nil {
		return err
	}
	_, err = c.do(ctx, "SET", key, value, "PX", ms)

	return err
}

func (c *Client) SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	ms, err := ttlMillis(ttl)
	if err != nil {
		return false, err
	}
	_, err = redigo.String(c.do(ctx, "SET", key, value, "NX", "PX", ms))
	if errors.Is(err, redigo.ErrNil) {
		return false, nil // ключ уже существует
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := redigo.Bytes(c.do(ctx, "GET", key))
	if errors.Is(err, redigo.ErrNil) {
		return nil, cache.ErrNotFound
	}

	return b, err
}

func (c *Client) Del(ctx context.Context, key string) error {
	_, err := c.do(ctx, "DEL", key)
	return err
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	return redigo.Bool(c.do(ctx, "EXISTS", key))
}

func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	ms, err := ttlMillis(ttl)
	if err != nil {
		return err
	}
	_, err = c.do(ctx, "PEXPIRE", key, ms)

	return err
}

func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return redigo.Int64(c.do(ctx, "INCR", key))
}

func (c *Client) HashSet(ctx context.Context, key string, values any) error {
	_, err := c.do(ctx, "HSET", redigo.Args{key}.AddFlat(values)...)
	return err
}

func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return redigo.StringMap(c.do(ctx, "HGETALL", key))
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.do(ctx, "PING")
	return err
}

func (c *Client) Publish(ctx context.Context, channel string, payload []byte) error {
	_, err := c.do(ctx, "PUBLISH", channel, payload)
	return err
}
