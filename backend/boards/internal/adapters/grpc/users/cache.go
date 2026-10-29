package users

import (
	"sync"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
)

// maxCacheEntries предел размера кеша. При переполнении кеш сбрасывается целиком: проще и надёжнее вытеснения,
// а заново профили подтянутся по мере запросов.
const maxCacheEntries = 10_000

type cacheEntry struct {
	user    models.User
	expires time.Time
}

// cache профили по id с общим временем жизни записи. Безопасен для параллельного использования.
type cache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
	ttl   time.Duration
	now   func() time.Time
}

func newCache(ttl time.Duration) *cache {
	return &cache{items: map[string]cacheEntry{}, ttl: ttl, now: time.Now}
}

func (c *cache) get(id string) (models.User, bool) {
	if c.ttl <= 0 {
		return models.User{}, false
	}

	c.mu.RLock()
	e, ok := c.items[id]
	c.mu.RUnlock()

	if !ok || !c.now().Before(e.expires) {
		return models.User{}, false
	}

	return e.user, true
}

func (c *cache) put(u models.User) {
	if c.ttl <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.items) >= maxCacheEntries {
		c.items = make(map[string]cacheEntry, maxCacheEntries/2)
	}
	c.items[u.ID] = cacheEntry{user: u, expires: c.now().Add(c.ttl)}
}
