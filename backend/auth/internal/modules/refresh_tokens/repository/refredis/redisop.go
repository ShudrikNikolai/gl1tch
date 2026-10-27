package refredis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens/repository/converter"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens/repository/model"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/cache"
)

func (r *Repo) Save(ctx context.Context, hash string, s refresh_tokens.Session, ttl time.Duration) error {
	raw, err := json.Marshal(converter.SessionSrvToModel(s))
	if err != nil {
		return fmt.Errorf("refresh_tokens: marshal session: %w", err)
	}

	if err = r.rdb.SetWithTTL(ctx, keySession(hash), raw, ttl); err != nil {
		return fmt.Errorf("refresh_tokens: save session: %w", err)
	}
	if err = r.rdb.SAdd(ctx, keyUser(s.UserID), hash); err != nil {
		return fmt.Errorf("refresh_tokens: index session: %w", err)
	}

	if err = r.rdb.Expire(ctx, keyUser(s.UserID), ttl); err != nil {
		r.log.WarnContext(ctx, "refresh_tokens: expire user set", "error", err)
	}

	return nil
}

func (r *Repo) Get(ctx context.Context, hash string) (refresh_tokens.Session, error) {
	raw, err := r.rdb.Get(ctx, keySession(hash))
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return refresh_tokens.Session{}, refresh_tokens.ErrSessionNotFound
		}

		return refresh_tokens.Session{}, fmt.Errorf("refresh_tokens: get session: %w", err)
	}

	var m model.Session
	if err = json.Unmarshal(raw, &m); err != nil {
		return refresh_tokens.Session{}, fmt.Errorf("refresh_tokens: decode session: %w", err)
	}

	return converter.SessionModelToSrv(m), nil
}

func (r *Repo) Delete(ctx context.Context, hash, userID string) error {
	if err := r.rdb.Del(ctx, keySession(hash)); err != nil {
		return fmt.Errorf("refresh_tokens: delete session: %w", err)
	}
	if userID != "" {
		if err := r.rdb.SRem(ctx, keyUser(userID), hash); err != nil {
			return fmt.Errorf("refresh_tokens: unindex session: %w", err)
		}
	}

	return nil
}

func (r *Repo) DeleteAllByUser(ctx context.Context, userID string) error {
	hashes, err := r.rdb.SMembers(ctx, keyUser(userID))
	if err != nil {
		return fmt.Errorf("refresh_tokens: list sessions: %w", err)
	}

	for _, h := range hashes {
		if err = r.rdb.Del(ctx, keySession(h)); err != nil {
			return fmt.Errorf("refresh_tokens: delete session: %w", err)
		}
	}
	if err = r.rdb.Del(ctx, keyUser(userID)); err != nil {
		return fmt.Errorf("refresh_tokens: delete user index: %w", err)
	}

	return nil
}

func (r *Repo) MarkUsed(ctx context.Context, hash, userID string, at time.Time, ttl time.Duration) (bool, error) {
	val := userID + "|" + strconv.FormatInt(at.UnixMilli(), 10)

	ok, err := r.rdb.SetNX(ctx, keyUsed(hash), val, ttl)
	if err != nil {
		return false, fmt.Errorf("refresh_tokens: mark used: %w", err)
	}

	return ok, nil
}

func (r *Repo) Used(ctx context.Context, hash string) (string, time.Time, bool, error) {
	raw, err := r.rdb.Get(ctx, keyUsed(hash))
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return "", time.Time{}, false, nil
		}

		return "", time.Time{}, false, fmt.Errorf("refresh_tokens: get used: %w", err)
	}

	userID, ms, ok := strings.Cut(string(raw), "|")
	if !ok {
		return "", time.Time{}, false, fmt.Errorf("refresh_tokens: malformed used mark %q", raw)
	}
	n, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return "", time.Time{}, false, fmt.Errorf("refresh_tokens: malformed used mark %q: %w", raw, err)
	}

	return userID, time.UnixMilli(n), true, nil
}
