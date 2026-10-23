package jwtx

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/cache"
)

// TODO: перенести потом в redis
// Revocations локальный в памяти «blacklist»: для пользователя хранится момент, раньше которого
// выданные токены недействительны (бан, смена пароля, выход со всех устройств). Одна запись на пользователя,
// а не на токен. Обновления приходят по pub/sub (Run), поэтому проверка на каждом запросе не ходит в сеть.
//
// Ограничение: после рестарта процесса список пуст, и токен, отозванный за последние минуты (до exp),
// будет принят. Для строгости при старте можно подгружать отзывы из источника правды (auth).
type Revocations struct {
	mu     sync.RWMutex
	before map[string]time.Time
}

func NewRevocations() *Revocations {
	return &Revocations{before: make(map[string]time.Time)}
}

// Revoke запоминает, что токены пользователя, выданные раньше t, недействительны. Старшую отметку не затирает.
func (r *Revocations) Revoke(userID string, t time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cur, ok := r.before[userID]; !ok || t.After(cur) {
		r.before[userID] = t
	}
}

// IsRevoked: токен, выданный в момент issuedAt, отозван?
func (r *Revocations) IsRevoked(userID string, issuedAt time.Time) bool {
	r.mu.RLock()
	t, ok := r.before[userID]
	r.mu.RUnlock()

	return ok && issuedAt.Before(t)
}

// Prune удаляет отметки старше maxAge (токены, выданные до них, уже истекли сами). Вызывайте периодически.
func (r *Revocations) Prune(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, t := range r.before {
		if t.Before(cutoff) {
			delete(r.before, id)
		}
	}
}

// PublishRevocation отправляет отзыв всем экземплярам gateway. Формат сообщения: "<userID>|<unix_millis>".
func PublishRevocation(ctx context.Context, ps cache.PubSub, channel, userID string, before time.Time) error {
	return ps.Publish(ctx, channel, []byte(userID+"|"+strconv.FormatInt(before.UnixMilli(), 10)))
}

// Run подписывается на канал и применяет отзывы, пока ctx не отменён. Подписка переподключается сама.
func (r *Revocations) Run(ctx context.Context, ps cache.PubSub, channel string, log *slog.Logger) error {
	ch, err := ps.Subscribe(ctx, channel)
	if err != nil {
		return err
	}
	for m := range ch {
		id, ts, ok := strings.Cut(string(m.Payload), "|")
		ms, perr := strconv.ParseInt(ts, 10, 64)
		if !ok || id == "" || perr != nil {
			log.WarnContext(ctx, "jwtx: malformed revocation message", slog.String("payload", string(m.Payload)))
			continue
		}
		r.Revoke(id, time.UnixMilli(ms))
	}

	return ctx.Err()
}
