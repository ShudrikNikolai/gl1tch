package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/uuid"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/credentials"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/credentials/repository"
)

// Hasher хеширование паролей (реализация: infrastructure/password.Argon2).
type Hasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Verify(ctx context.Context, password, encoded string) (ok, needsRehash bool, err error)
}

type Service struct {
	repo   repository.Repository
	hasher Hasher
	log    *slog.Logger

	// dummyHash сверяется, когда логина нет: время ответа не выдаёт, существует ли пользователь.
	dummyHash string
}

func New(repo repository.Repository, hasher Hasher, log *slog.Logger) (*Service, error) {
	dummy, err := hasher.Hash(context.Background(), "gl1tch-timing-equalizer")
	if err != nil {
		return nil, fmt.Errorf("credentials: dummy hash: %w", err)
	}

	return &Service{repo: repo, hasher: hasher, log: log, dummyHash: dummy}, nil
}

// Create заводит учётку с новым user_id (UUIDv7). Логин должен быть уже провалидирован.
func (s *Service) Create(ctx context.Context, username, password string) (credentials.Credential, error) {
	hash, err := s.hasher.Hash(ctx, password)
	if err != nil {
		return credentials.Credential{}, fmt.Errorf("credentials: hash: %w", err)
	}

	c := credentials.Credential{
		UserID:       uuid.NewUUIDv7(time.Now()),
		Username:     username,
		PasswordHash: hash,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return credentials.Credential{}, err
	}

	return c, nil
}

// Authenticate проверяет пару логин/пароль. Любая неудача для клиента выглядит одинаково (ErrInvalidCredentials).
func (s *Service) Authenticate(ctx context.Context, username, password string) (credentials.Credential, error) {
	c, err := s.repo.GetByUsername(ctx, username)
	if errors.Is(err, credentials.ErrNotFound) {
		_, _, _ = s.hasher.Verify(ctx, password, s.dummyHash)

		return credentials.Credential{}, credentials.ErrInvalidCredentials
	}
	if err != nil {
		return credentials.Credential{}, err
	}

	ok, needsRehash, err := s.hasher.Verify(ctx, password, c.PasswordHash)
	if err != nil {
		if ctx.Err() != nil {
			return credentials.Credential{}, err
		}
		// битый хеш в БД: клиенту это неотличимо от неверного пароля, в лог пишем причину
		s.log.ErrorContext(ctx, "credentials: verify failed", slog.String("user_id", c.UserID), slog.Any("error", err))

		return credentials.Credential{}, credentials.ErrInvalidCredentials
	}
	if !ok {
		return credentials.Credential{}, credentials.ErrInvalidCredentials
	}

	if needsRehash {
		s.rehash(ctx, &c, password)
	}

	return c, nil
}

// rehash обновляет хеш под текущие параметры Argon2. Ошибка не мешает входу.
func (s *Service) rehash(ctx context.Context, c *credentials.Credential, password string) {
	hash, err := s.hasher.Hash(ctx, password)
	if err == nil {
		err = s.repo.UpdateHash(ctx, c.UserID, hash)
	}
	if err != nil {
		s.log.WarnContext(ctx, "credentials: rehash failed", slog.String("user_id", c.UserID), slog.Any("error", err))
		return
	}
	c.PasswordHash = hash
}

// Delete удаляет учётку (откат неудачной регистрации).
func (s *Service) Delete(ctx context.Context, userID string) error {
	return s.repo.Delete(ctx, userID)
}
