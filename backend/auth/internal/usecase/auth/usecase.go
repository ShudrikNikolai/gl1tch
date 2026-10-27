package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/credentials"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/models"
)

var _ usecase.AuthUseCase = (*UseCase)(nil)

// Интерфейсы объявлены здесь, у потребителя. Реализации: modules/*/service, adapters/grpc/users, jwtx.Signer.

type CredentialsService interface {
	Create(ctx context.Context, username, password string) (credentials.Credential, error)
	Authenticate(ctx context.Context, username, password string) (credentials.Credential, error)
	Delete(ctx context.Context, userID string) error
}

type RefreshTokenService interface {
	Issue(ctx context.Context, userID string) (refresh_tokens.Token, error)
	Rotate(ctx context.Context, raw string) (userID string, next refresh_tokens.Token, err error)
	Revoke(ctx context.Context, raw string) error
	RevokeAll(ctx context.Context, userID string) error
}

type UsersClient interface {
	CreateUser(ctx context.Context, id, username string) (models.User, error)
	GetUser(ctx context.Context, id string) (models.User, error)
}

type AccessSigner interface {
	Sign(userID string, roles []string, now time.Time) (token string, expiresAt time.Time, err error)
}

type Deps struct {
	Credentials CredentialsService
	Refresh     RefreshTokenService
	Users       UsersClient
	Signer      AccessSigner
	Log         *slog.Logger
}

type UseCase struct {
	creds   CredentialsService
	refresh RefreshTokenService
	users   UsersClient
	signer  AccessSigner
	log     *slog.Logger
	now     func() time.Time
}

func New(d Deps) *UseCase {
	return &UseCase{
		creds:   d.Credentials,
		refresh: d.Refresh,
		users:   d.Users,
		signer:  d.Signer,
		log:     d.Log,
		now:     time.Now,
	}
}

// accessToken подписывает access-токен; роли берутся из профиля в users.
func (uc *UseCase) accessToken(u models.User) (string, time.Time, error) {
	return uc.signer.Sign(u.ID, []string{u.Role}, uc.now())
}

// issueTokens выдаёт пару токенов (для register и login).
func (uc *UseCase) issueTokens(ctx context.Context, u models.User) (models.TokenPair, error) {
	refresh, err := uc.refresh.Issue(ctx, u.ID)
	if err != nil {
		return models.TokenPair{}, err
	}

	return uc.pair(u, refresh)
}

// pair собирает пару из уже выданного refresh-токена.
func (uc *UseCase) pair(u models.User, refresh refresh_tokens.Token) (models.TokenPair, error) {
	access, accessExp, err := uc.accessToken(u)
	if err != nil {
		return models.TokenPair{}, err
	}

	return models.TokenPair{
		AccessToken:      access,
		RefreshToken:     refresh.Value,
		AccessExpiresAt:  &accessExp,
		RefreshExpiresAt: &refresh.ExpiresAt,
	}, nil
}
