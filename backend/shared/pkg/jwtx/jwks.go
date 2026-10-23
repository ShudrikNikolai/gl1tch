package jwtx

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// jwk один ключ из JWKS. Поддерживаются только Ed25519 (kty=OKP, crv=Ed25519): так подписывает auth.
type jwk struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	Kid string `json:"kid"`
	X   string `json:"x"`
	Use string `json:"use,omitempty"`
	Alg string `json:"alg,omitempty"`
}

type jwkSet struct {
	Keys []jwk `json:"keys"`
}

var (
	// ErrUnknownKey: в JWKS нет ключа с таким kid (чаще всего токен выпущен кем-то другим).
	ErrUnknownKey = errors.New("jwtx: unknown signing key")
	// ErrKeysUnavailable: ключи не удалось загрузить и кэш пуст.
	ErrKeysUnavailable = errors.New("jwtx: signing keys unavailable")
)

// jwksCache хранит публичные ключи, загруженные по URL. Обновляется:
//   - при неизвестном kid (ротация ключа), но не чаще minGap;
//   - по истечении refreshEvery (при очередном обращении).
//
// Если обновление не удалось, а в кэше уже есть ключи, используются старые (с предупреждением в лог).
type jwksCache struct {
	url          string
	client       *http.Client
	log          *slog.Logger
	refreshEvery time.Duration
	minGap       time.Duration

	mu          sync.Mutex
	keys        map[string]ed25519.PublicKey
	lastAttempt time.Time
	lastSuccess time.Time
}

func newJWKSCache(url string, client *http.Client, log *slog.Logger, refreshEvery, minGap time.Duration) *jwksCache {
	return &jwksCache{url: url, client: client, log: log, refreshEvery: refreshEvery, minGap: minGap}
}

// key возвращает ключ по kid, при необходимости обновляя кэш.
func (c *jwksCache) key(ctx context.Context, kid string) (ed25519.PublicKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if k, ok := c.keys[kid]; ok && now.Sub(c.lastSuccess) < c.refreshEvery {
		return k, nil
	}

	// кэш пуст, устарел или kid неизвестен: пробуем обновить, не чаще minGap
	if c.keys == nil || now.Sub(c.lastAttempt) >= c.minGap {
		if err := c.refreshLocked(ctx); err != nil {
			if c.keys == nil {
				return nil, fmt.Errorf("%w: %w", ErrKeysUnavailable, err)
			}
			c.log.WarnContext(ctx, "jwtx: jwks refresh failed, using cached keys", slog.Any("error", err))
		}
	}

	if k, ok := c.keys[kid]; ok {
		return k, nil
	}

	return nil, ErrUnknownKey
}

func (c *jwksCache) refreshLocked(ctx context.Context) error {
	c.lastAttempt = time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks endpoint returned %s", resp.Status)
	}

	var set jwkSet
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&set); err != nil {
		return fmt.Errorf("decode jwks: %w", err)
	}

	keys := make(map[string]ed25519.PublicKey, len(set.Keys))
	for _, k := range set.Keys {
		if k.Kty != "OKP" || k.Crv != "Ed25519" || k.Kid == "" {
			continue // чужие типы ключей пропускаем
		}
		raw, err := base64.RawURLEncoding.DecodeString(k.X)
		if err != nil || len(raw) != ed25519.PublicKeySize {
			continue
		}
		keys[k.Kid] = ed25519.PublicKey(raw)
	}
	if len(keys) == 0 {
		return errors.New("jwks contains no usable Ed25519 keys")
	}

	c.keys = keys
	c.lastSuccess = time.Now()

	return nil
}
