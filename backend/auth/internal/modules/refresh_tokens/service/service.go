package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens/repository"
)

// Revoker оповещает остальные сервисы, что access-токены пользователя, выданные раньше before, недействительны.
type Revoker interface {
	RevokeUser(ctx context.Context, userID string, before time.Time) error
}

type Service struct {
	repo    repository.Repository
	revoker Revoker
	ttl     time.Duration // время жизни refresh-токена
	grace   time.Duration // окно, в котором повторное предъявление считается гонкой, а не кражей
	log     *slog.Logger
	now     func() time.Time
}

func New(repo repository.Repository, revoker Revoker, ttl, grace time.Duration, log *slog.Logger) *Service {
	return &Service{repo: repo, revoker: revoker, ttl: ttl, grace: grace, log: log, now: time.Now}
}

// Issue выдаёт новый refresh-токен.
func (s *Service) Issue(ctx context.Context, userID string) (refresh_tokens.Token, error) {
	raw, err := newToken()
	if err != nil {
		return refresh_tokens.Token{}, err
	}

	now := s.now()
	if err := s.repo.Save(ctx, hashToken(raw), refresh_tokens.Session{UserID: userID, IssuedAt: now}, s.ttl); err != nil {
		return refresh_tokens.Token{}, err
	}

	return refresh_tokens.Token{Value: raw, ExpiresAt: now.Add(s.ttl)}, nil
}

// Rotate обменивает токен на новый (старый перестаёт работать) и возвращает владельца.
//
// Повторное предъявление уже использованного токена после окна grace считается кражей:
// отзываются все сессии пользователя и его access-токены.
func (s *Service) Rotate(ctx context.Context, raw string) (string, refresh_tokens.Token, error) {
	if raw == "" {
		return "", refresh_tokens.Token{}, refresh_tokens.ErrInvalidToken
	}
	hash := hashToken(raw)

	sess, err := s.repo.Get(ctx, hash)
	if errors.Is(err, refresh_tokens.ErrSessionNotFound) {
		return "", refresh_tokens.Token{}, s.rejectUnknown(ctx, hash)
	}
	if err != nil {
		return "", refresh_tokens.Token{}, err
	}

	// Из параллельных запросов с одним токеном выигрывает ровно один.
	won, err := s.repo.MarkUsed(ctx, hash, sess.UserID, s.now(), s.ttl)
	if err != nil {
		return "", refresh_tokens.Token{}, err
	}
	if !won {
		return "", refresh_tokens.Token{}, s.rejectUnknown(ctx, hash)
	}

	next, err := s.Issue(ctx, sess.UserID)
	if err != nil {
		return "", refresh_tokens.Token{}, err
	}
	if err := s.repo.Delete(ctx, hash, sess.UserID); err != nil {
		s.log.WarnContext(ctx, "refresh_tokens: delete rotated token", slog.Any("error", err))
	}

	return sess.UserID, next, nil
}

// rejectUnknown решает, чем считать токен, которого нет или который уже использован.
func (s *Service) rejectUnknown(ctx context.Context, hash string) error {
	userID, at, found, err := s.repo.Used(ctx, hash)
	if err != nil {
		return err
	}
	if !found || s.now().Sub(at) <= s.grace {
		return refresh_tokens.ErrInvalidToken
	}

	s.log.WarnContext(ctx, "refresh_tokens: reuse detected, revoking all sessions", slog.String("user_id", userID))
	if err := s.RevokeAll(ctx, userID); err != nil {
		s.log.ErrorContext(ctx, "refresh_tokens: revoke after reuse failed", slog.String("user_id", userID), slog.Any("error", err))
	}

	return refresh_tokens.ErrTokenReused
}

// Revoke завершает одну сессию (logout). Неизвестный токен не ошибка.
func (s *Service) Revoke(ctx context.Context, raw string) error {
	if raw == "" {
		return nil
	}
	hash := hashToken(raw)

	sess, err := s.repo.Get(ctx, hash)
	if errors.Is(err, refresh_tokens.ErrSessionNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, hash, sess.UserID)
}

// RevokeAll завершает все сессии пользователя и отзывает его уже выданные access-токены.
func (s *Service) RevokeAll(ctx context.Context, userID string) error {
	if err := s.repo.DeleteAllByUser(ctx, userID); err != nil {
		return err
	}

	return s.revoker.RevokeUser(ctx, userID, s.now())
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("refresh_tokens: random: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(sum[:])
}
