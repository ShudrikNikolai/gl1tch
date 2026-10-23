package password

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/config"
)

// малые параметры: тесты не должны считать по 19 МиБ
var fast = config.Argon2{MemoryKiB: 64, Iterations: 1, Parallelism: 1, MaxConcurrent: 2}

func TestHashVerify(t *testing.T) {
	a := NewArgon2(fast)
	ctx := context.Background()

	h, err := a.Hash(ctx, "correct horse")
	if err != nil || !strings.HasPrefix(h, "$argon2id$v=19$m=64,t=1,p=1$") {
		t.Fatalf("hash: %q %v", h, err)
	}
	ok, rehash, err := a.Verify(ctx, "correct horse", h)
	if err != nil || !ok || rehash {
		t.Fatalf("verify: ok=%v rehash=%v err=%v", ok, rehash, err)
	}
	if ok, _, _ := a.Verify(ctx, "wrong", h); ok {
		t.Fatal("wrong password must not verify")
	}
}

func TestSaltsAreUnique(t *testing.T) {
	a := NewArgon2(fast)
	h1, _ := a.Hash(context.Background(), "same")
	h2, _ := a.Hash(context.Background(), "same")
	if h1 == h2 {
		t.Fatal("same password must produce different hashes")
	}
}

func TestNeedsRehashWhenParamsChange(t *testing.T) {
	old := NewArgon2(fast)
	h, _ := old.Hash(context.Background(), "pw")

	stronger := fast
	stronger.MemoryKiB = 128
	ok, rehash, err := NewArgon2(stronger).Verify(context.Background(), "pw", h)
	if err != nil || !ok || !rehash {
		t.Fatalf("old hash must still verify and request a rehash: ok=%v rehash=%v err=%v", ok, rehash, err)
	}
}

func TestMalformedHashes(t *testing.T) {
	a := NewArgon2(fast)
	for _, bad := range []string{
		"", "plain", "$argon2i$v=19$m=64,t=1,p=1$c2FsdA$a2V5",
		"$argon2id$v=19$m=64,t=1,p=1$!!!$a2V5",
		"$argon2id$v=19$m=99999999,t=1,p=1$c2FsdA$a2V5", // слишком прожорливая хуйня
		"$argon2id$v=18$m=64,t=1,p=1$c2FsdA$a2V5",
	} {
		if _, _, err := a.Verify(context.Background(), "x", bad); !errors.Is(err, ErrMalformedHash) {
			t.Errorf("%q: want ErrMalformedHash, got %v", bad, err)
		}
	}
}

func TestConcurrencyLimitHonorsContext(t *testing.T) {
	cnf := fast
	cnf.MaxConcurrent = 1
	a := NewArgon2(cnf)
	a.sem <- struct{}{} // слот занят
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, err := a.Hash(ctx, "x"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("waiting for a slot must respect ctx, got %v", err)
	}
}
