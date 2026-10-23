package jwtx

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/cache"
)

var discard = slog.New(slog.NewTextHandler(io.Discard, nil))

const issuer = "gl1tch-auth"

func newSigner(t *testing.T) *Signer {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	return NewSigner(priv, issuer, 5*time.Minute)
}

// jwksServer раздаёт JWKS текущего signer и считает обращения.
type jwksServer struct {
	*httptest.Server
	mu     sync.Mutex
	signer *Signer
	status int
	hits   atomic.Int32
}

func newJWKSServer(t *testing.T, s *Signer) *jwksServer {
	t.Helper()
	js := &jwksServer{signer: s, status: http.StatusOK}
	js.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		js.hits.Add(1)
		js.mu.Lock()
		defer js.mu.Unlock()
		if js.status != http.StatusOK {
			w.WriteHeader(js.status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(js.signer.JWKS())
	}))
	t.Cleanup(js.Close)

	return js
}

func (j *jwksServer) set(s *Signer, status int) {
	j.mu.Lock()
	j.signer, j.status = s, status
	j.mu.Unlock()
}

func verifier(url string, mutate ...func(*VerifierConfig)) *Verifier {
	cfg := VerifierConfig{JWKSURL: url, Issuer: issuer, RefreshInterval: time.Hour, MinRefreshGap: time.Millisecond,
		HTTPTimeout: 2 * time.Second, Leeway: time.Second}
	for _, m := range mutate {
		m(&cfg)
	}

	return NewVerifier(cfg, discard)
}

func TestSignVerifyRoundTrip(t *testing.T) {
	s := newSigner(t)
	srv := newJWKSServer(t, s)
	v := verifier(srv.URL)

	token, exp, err := s.Sign("user-1", []string{"user", "moder"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	claims, err := v.Verify(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "user-1" || len(claims.Roles) != 2 || claims.Roles[1] != "moder" || claims.Issuer != issuer || claims.ID == "" {
		t.Fatalf("claims: %+v", claims)
	}
	if d := claims.ExpiresAt.Time.Sub(exp); d < -time.Second || d > time.Second {
		t.Fatalf("exp mismatch: %v", d)
	}
}

func TestExpiredToken(t *testing.T) {
	s := newSigner(t)
	srv := newJWKSServer(t, s)
	token, _, _ := s.Sign("u", nil, time.Now().Add(-time.Hour))
	if _, err := verifier(srv.URL).Verify(context.Background(), token); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("want ErrExpiredToken, got %v", err)
	}
}

func TestWrongIssuerAndTampering(t *testing.T) {
	s := newSigner(t)
	srv := newJWKSServer(t, s)
	ctx := context.Background()

	token, _, _ := s.Sign("u", []string{"user"}, time.Now())
	if _, err := verifier(srv.URL, func(c *VerifierConfig) { c.Issuer = "someone-else" }).Verify(ctx, token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("wrong issuer: %v", err)
	}

	// подмена полезной нагрузки (повышение роли) ломает подпись
	parts := strings.Split(token, ".")
	payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
	forged := strings.Replace(string(payload), `"user"`, `"moder"`, 1)
	parts[1] = base64.RawURLEncoding.EncodeToString([]byte(forged))
	if _, err := verifier(srv.URL).Verify(ctx, strings.Join(parts, ".")); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("tampered payload must be rejected, got %v", err)
	}
}

func TestRejectsOtherAlgorithms(t *testing.T) {
	s := newSigner(t)
	srv := newJWKSServer(t, s)
	v := verifier(srv.URL)
	claims := jwt.MapClaims{"sub": "u", "iss": issuer, "exp": time.Now().Add(time.Hour).Unix()}

	// HS256 с произвольным секретом и верным kid
	hs := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	hs.Header["kid"] = s.KeyID()
	hsTok, _ := hs.SignedString([]byte("secret"))
	if _, err := v.Verify(context.Background(), hsTok); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("HS256 must be rejected: %v", err)
	}

	// alg=none
	none := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	none.Header["kid"] = s.KeyID()
	noneTok, _ := none.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if _, err := v.Verify(context.Background(), noneTok); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("alg=none must be rejected: %v", err)
	}

	// токен без kid
	noKid := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	noKidTok, _ := noKid.SignedString(priv)
	if _, err := v.Verify(context.Background(), noKidTok); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("token without kid must be rejected: %v", err)
	}
}

func TestKeyRotationTriggersRefresh(t *testing.T) {
	oldS, newS := newSigner(t), newSigner(t)
	srv := newJWKSServer(t, oldS)
	v := verifier(srv.URL)
	ctx := context.Background()

	if err := v.Warmup(ctx); err != nil {
		t.Fatal(err)
	}
	srv.set(newS, http.StatusOK) // auth начал подписывать новым ключом
	time.Sleep(5 * time.Millisecond)

	token, _, _ := newS.Sign("u", nil, time.Now())
	if _, err := v.Verify(ctx, token); err != nil {
		t.Fatalf("token signed with the rotated key must verify after refresh: %v", err)
	}
}

func TestUnknownKidDoesNotHammerJWKS(t *testing.T) {
	s := newSigner(t)
	srv := newJWKSServer(t, s)
	v := verifier(srv.URL, func(c *VerifierConfig) { c.MinRefreshGap = time.Hour })

	stranger := newSigner(t)
	token, _, _ := stranger.Sign("u", nil, time.Now())
	for i := 0; i < 50; i++ {
		if _, err := v.Verify(context.Background(), token); !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("unknown kid must be an invalid token: %v", err)
		}
	}
	if hits := srv.hits.Load(); hits != 1 {
		t.Fatalf("jwks endpoint was hit %d times, want 1", hits)
	}
}

func TestStaleKeysUsedWhenRefreshFails(t *testing.T) {
	s := newSigner(t)
	srv := newJWKSServer(t, s)
	v := verifier(srv.URL, func(c *VerifierConfig) { c.RefreshInterval = time.Millisecond })
	if err := v.Warmup(context.Background()); err != nil {
		t.Fatal(err)
	}
	srv.set(s, http.StatusInternalServerError)
	time.Sleep(5 * time.Millisecond)

	token, _, _ := s.Sign("u", nil, time.Now())
	if _, err := v.Verify(context.Background(), token); err != nil {
		t.Fatalf("cached key must keep working while auth is down: %v", err)
	}
}

func TestKeysUnavailableIsNotInvalidToken(t *testing.T) {
	s := newSigner(t)
	srv := newJWKSServer(t, s)
	srv.set(s, http.StatusServiceUnavailable)
	token, _, _ := s.Sign("u", nil, time.Now())

	_, err := verifier(srv.URL).Verify(context.Background(), token)
	if !errors.Is(err, ErrKeysUnavailable) || errors.Is(err, ErrInvalidToken) {
		t.Fatalf("infrastructure failure must not look like a bad token: %v", err)
	}
}

func TestJWKSFormat(t *testing.T) {
	s := newSigner(t)
	var set struct {
		Keys []map[string]string `json:"keys"`
	}
	if err := json.Unmarshal(s.JWKS(), &set); err != nil || len(set.Keys) != 1 {
		t.Fatalf("jwks: %v %s", err, s.JWKS())
	}
	k := set.Keys[0]
	if k["kty"] != "OKP" || k["crv"] != "Ed25519" || k["kid"] != s.KeyID() || k["alg"] != "EdDSA" || k["x"] == "" {
		t.Fatalf("jwk: %v", k)
	}
}

func writePEM(t *testing.T, blockType string, der []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "key.pem")
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestLoadSigner(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	der, _ := x509.MarshalPKCS8PrivateKey(priv)
	s, err := LoadSigner(writePEM(t, "PRIVATE KEY", der), issuer, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	srv := newJWKSServer(t, s)
	token, _, _ := s.Sign("u", nil, time.Now())
	if _, err := verifier(srv.URL).Verify(context.Background(), token); err != nil {
		t.Fatal(err)
	}

	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	rsaDER, _ := x509.MarshalPKCS8PrivateKey(rsaKey)
	if _, err := LoadSigner(writePEM(t, "PRIVATE KEY", rsaDER), issuer, time.Minute); err == nil {
		t.Fatal("RSA key must be rejected")
	}
	if _, err := LoadSigner(filepath.Join(t.TempDir(), "missing.pem"), issuer, time.Minute); err == nil {
		t.Fatal("missing file must be an error")
	}
	bad := filepath.Join(t.TempDir(), "bad.pem")
	_ = os.WriteFile(bad, []byte("not pem"), 0o600)
	if _, err := LoadSigner(bad, issuer, time.Minute); err == nil {
		t.Fatal("garbage must be an error")
	}
}

func TestRevocations(t *testing.T) {
	r := NewRevocations()
	issued := time.Now().Add(-time.Minute)
	if r.IsRevoked("u", issued) {
		t.Fatal("nothing revoked yet")
	}
	r.Revoke("u", time.Now())
	if !r.IsRevoked("u", issued) {
		t.Fatal("token issued before the cutoff must be revoked")
	}
	if r.IsRevoked("u", time.Now().Add(time.Second)) {
		t.Fatal("token issued after the cutoff must pass (re-login)")
	}
	if r.IsRevoked("other", issued) {
		t.Fatal("other users.yaml are unaffected")
	}

	r.Revoke("u", time.Now().Add(-time.Hour)) // старая отметка не затирает новую
	if !r.IsRevoked("u", issued) {
		t.Fatal("older cutoff must not overwrite a newer one")
	}

	r.Revoke("old", time.Now().Add(-2*time.Hour))
	r.Prune(time.Hour)
	if r.IsRevoked("old", time.Now().Add(-3*time.Hour)) {
		t.Fatal("expired cutoff must be pruned")
	}
	if !r.IsRevoked("u", issued) {
		t.Fatal("fresh cutoff must survive pruning")
	}
}

type fakePubSub struct {
	ch  chan cache.Message
	pub []cache.Message
}

func (f *fakePubSub) Publish(_ context.Context, channel string, payload []byte) error {
	f.pub = append(f.pub, cache.Message{Channel: channel, Payload: payload})
	return nil
}

func (f *fakePubSub) Subscribe(context.Context, ...string) (<-chan cache.Message, error) {
	return f.ch, nil
}

func TestRevocationsRunAppliesMessages(t *testing.T) {
	ps := &fakePubSub{ch: make(chan cache.Message, 4)}
	r := NewRevocations()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- r.Run(ctx, ps, "revoked", discard) }()

	cutoff := time.Now()
	if err := PublishRevocation(ctx, ps, "revoked", "user-9", cutoff); err != nil {
		t.Fatal(err)
	}
	ps.ch <- ps.pub[0]                                       // доставка опубликованного
	ps.ch <- cache.Message{Payload: []byte("garbage")}       // битое сообщение пропускается
	ps.ch <- cache.Message{Payload: []byte("user-9|notnum")} // и это тоже

	deadline := time.After(2 * time.Second)
	for !r.IsRevoked("user-9", cutoff.Add(-time.Second)) {
		select {
		case <-deadline:
			t.Fatal("revocation was not applied")
		case <-time.After(5 * time.Millisecond):
		}
	}

	cancel()
	close(ps.ch)
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run returned %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop")
	}
}
