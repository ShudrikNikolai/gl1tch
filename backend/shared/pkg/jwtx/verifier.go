package jwtx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims claims access-токена. Subject это id пользователя.
type Claims struct {
	jwt.RegisteredClaims
	Roles    []string `json:"roles"`
	RolesCSV string   `json:"roles_csv,omitempty"` // для Envoy: claim_to_headers умеет только строки
}

var (
	ErrInvalidToken = errors.New("jwtx: invalid token")
	ErrExpiredToken = errors.New("jwtx: token expired")
)

// VerifierConfig читается из окружения.
type VerifierConfig struct {
	JWKSURL         string        `env:"JWKS_URL,required"`
	Issuer          string        `env:"JWT_ISSUER"           envDefault:"gl1tch-auth"`
	RefreshInterval time.Duration `env:"JWKS_REFRESH_INTERVAL" envDefault:"10m"`
	MinRefreshGap   time.Duration `env:"JWKS_MIN_REFRESH_GAP"  envDefault:"10s"`
	HTTPTimeout     time.Duration `env:"JWKS_HTTP_TIMEOUT"     envDefault:"5s"`
	Leeway          time.Duration `env:"JWT_LEEWAY"            envDefault:"5s"`
}

type Verifier struct {
	jwks   *jwksCache
	issuer string
	leeway time.Duration
}

func NewVerifier(cfg VerifierConfig, log *slog.Logger) *Verifier {
	if log == nil {
		log = slog.Default()
	}

	return &Verifier{
		jwks:   newJWKSCache(cfg.JWKSURL, &http.Client{Timeout: cfg.HTTPTimeout}, log, cfg.RefreshInterval, cfg.MinRefreshGap),
		issuer: cfg.Issuer,
		leeway: cfg.Leeway,
	}
}

// Warmup загружает ключи заранее (например, при старте), чтобы первый запрос не ждал JWKS.
// Ошибку можно не считать фатальной: ключи будут загружены лениво при первой проверке.
func (v *Verifier) Warmup(ctx context.Context) error {
	v.jwks.mu.Lock()
	defer v.jwks.mu.Unlock()

	return v.jwks.refreshLocked(ctx)
}

// Verify проверяет подпись (только EdDSA), срок действия и издателя.
// Возвращает ErrExpiredToken, ErrInvalidToken (с причиной внутри) или ErrKeysUnavailable.
func (v *Verifier) Verify(ctx context.Context, token string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims,
		func(t *jwt.Token) (any, error) {
			kid, _ := t.Header["kid"].(string)
			if kid == "" {
				return nil, errors.New("missing kid")
			}

			return v.jwks.key(ctx, kid)
		},
		jwt.WithValidMethods([]string{"EdDSA"}),
		jwt.WithIssuer(v.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(v.leeway),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		// ошибка загрузки JWKS, gateway ответит 503, а не 401
		if errors.Is(err, ErrKeysUnavailable) {
			return nil, err
		}

		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}
	if claims.Subject == "" {
		return nil, fmt.Errorf("%w: empty subject", ErrInvalidToken)
	}

	return claims, nil
}
