package cache

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound возвращается, когда ключа нет. Детали драйвера наружу не утекают.
var ErrNotFound = errors.New("cache: key not found")

// Message сообщение pub/sub.
type Message struct {
	Channel string
	Payload []byte
}

type KV interface {
	Set(ctx context.Context, key string, value any) error
	SetWithTTL(ctx context.Context, key string, value any, ttl time.Duration) error
	// SetNX записывает значение, только если ключа ещё нет. true, если запись произошла.
	SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error)
	// Get возвращает ErrNotFound, если ключа нет.
	Get(ctx context.Context, key string) ([]byte, error)
	Del(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Incr(ctx context.Context, key string) (int64, error)
}

type Hash interface {
	HashSet(ctx context.Context, key string, values any) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
}

type SetOperator interface {
	SAdd(ctx context.Context, key, value string) error
	SRem(ctx context.Context, key, value string) error
	SIsMember(ctx context.Context, key, value string) (bool, error)
	SMembers(ctx context.Context, key string) ([]string, error)
}

// PubSub: доставка at-most-once (сообщения, опубликованные во время обрыва связи, теряются).
// Для отзыва токенов этого достаточно: источник правды хранится отдельно (ключи в Redis).
type PubSub interface {
	Publish(ctx context.Context, channel string, payload []byte) error
	// Subscribe возвращает канал сообщений. Подписка сама переподключается при обрыве;
	// канал закрывается, когда ctx отменён. К моменту возврата подписка уже активна.
	Subscribe(ctx context.Context, channels ...string) (<-chan Message, error)
}

type RedisClient interface {
	KV
	Hash
	SetOperator
	PubSub
	Ping(ctx context.Context) error
}
