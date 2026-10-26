package jwtx

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Signer подписывает access-токены (используется в auth).
type Signer struct {
	key    ed25519.PrivateKey
	kid    string
	issuer string
	ttl    time.Duration
}

// LoadSigner читает приватный ключ Ed25519 (PKCS#8 PEM), как его создаёт scripts/gen-secrets.mjs.
func LoadSigner(pemPath, issuer string, accessTTL time.Duration) (*Signer, error) {
	data, err := os.ReadFile(pemPath)
	if err != nil {
		return nil, fmt.Errorf("jwtx: read key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("jwtx: key file has no PEM block")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("jwtx: parse key: %w", err)
	}
	key, ok := parsed.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("jwtx: key is %T, want Ed25519", parsed)
	}

	return NewSigner(key, issuer, accessTTL), nil
}

func NewSigner(key ed25519.PrivateKey, issuer string, accessTTL time.Duration) *Signer {
	return &Signer{key: key, kid: keyID(key.Public().(ed25519.PublicKey)), issuer: issuer, ttl: accessTTL}
}

// keyID стабильный идентификатор ключа: первые 16 байт SHA-256 публичного ключа.
func keyID(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)

	return base64.RawURLEncoding.EncodeToString(sum[:16])
}

func (s *Signer) KeyID() string { return s.kid }

// Sign выпускает access-токен пользователя. now передаётся явно ради тестов.
func (s *Signer) Sign(userID string, roles []string, now time.Time) (token string, expiresAt time.Time, err error) {
	jti := make([]byte, 16)
	if _, err := rand.Read(jti); err != nil {
		return "", time.Time{}, err
	}
	expiresAt = now.Add(s.ttl)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        hex.EncodeToString(jti),
		},
		Roles:    roles,
		RolesCSV: strings.Join(roles, ","),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	t.Header["kid"] = s.kid
	token, err = t.SignedString(s.key)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("jwtx: sign: %w", err)
	}

	return token, expiresAt, nil
}

// JWKS возвращает JSON для эндпоинта /.well-known/jwks.json.
func (s *Signer) JWKS() []byte {
	pub := s.key.Public().(ed25519.PublicKey)
	out, _ := json.Marshal(jwkSet{Keys: []jwk{{
		Kty: "OKP", Crv: "Ed25519", Kid: s.kid, Use: "sig", Alg: "EdDSA",
		X: base64.RawURLEncoding.EncodeToString(pub),
	}}})

	return out
}
