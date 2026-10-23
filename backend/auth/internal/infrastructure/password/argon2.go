package password

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/config"
	"golang.org/x/crypto/argon2"
)

// Config параметры Argon2id. Значения по умолчанию соответствуют минимальным рекомендациям OWASP
// (19 MiB, 2 прохода, 1 поток). MaxConcurrent ограничивает число одновременных вычислений, потому что
// каждое занимает Memory КиБ: без лимита всплеск логинов съест память процесса.

const (
	saltLen = 16
	keyLen  = 32
)

var ErrMalformedHash = errors.New("password: malformed hash")

type Argon2 struct {
	cfg config.Argon2
	sem chan struct{}
}

func NewArgon2(cfg config.Argon2) *Argon2 {
	if cfg.MemoryKiB == 0 {
		cfg.MemoryKiB = 19456
	}
	if cfg.Iterations == 0 {
		cfg.Iterations = 2
	}
	if cfg.Parallelism == 0 {
		cfg.Parallelism = 1
	}
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = runtime.NumCPU()
	}

	return &Argon2{cfg: cfg, sem: make(chan struct{}, cfg.MaxConcurrent)}
}

func (a *Argon2) acquire(ctx context.Context) (release func(), err error) {
	select {
	case a.sem <- struct{}{}:
		return func() { <-a.sem }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Hash возвращает PHC-строку: $argon2id$v=19$m=...,t=...,p=...$<salt>$<hash>.
func (a *Argon2) Hash(ctx context.Context, password string) (string, error) {
	release, err := a.acquire(ctx)
	if err != nil {
		return "", err
	}
	defer release()

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("password: salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, a.cfg.Iterations, a.cfg.MemoryKiB, a.cfg.Parallelism, keyLen)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version,
		a.cfg.MemoryKiB, a.cfg.Iterations, a.cfg.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

// Verify сверяет пароль с хешем по параметрам, записанным в самом хеше (старые хеши продолжают работать).
// needsRehash=true, если параметры хеша отличаются от текущей конфигурации.
func (a *Argon2) Verify(ctx context.Context, password, encoded string) (ok, needsRehash bool, err error) {
	p, salt, want, err := decode(encoded)
	if err != nil {
		return false, false, err
	}

	release, err := a.acquire(ctx)
	if err != nil {
		return false, false, err
	}
	defer release()

	got := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, uint32(len(want)))
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return false, false, nil
	}

	needsRehash = p.memory != a.cfg.MemoryKiB || p.iterations != a.cfg.Iterations || p.parallelism != a.cfg.Parallelism

	return true, needsRehash, nil
}

type params struct {
	memory, iterations uint32
	parallelism        uint8
}

func decode(encoded string) (p params, salt, key []byte, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return p, nil, nil, ErrMalformedHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return p, nil, nil, ErrMalformedHash
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism); err != nil {
		return p, nil, nil, ErrMalformedHash
	}
	if p.memory == 0 || p.iterations == 0 || p.parallelism == 0 || p.memory > 1<<20 {
		return p, nil, nil, ErrMalformedHash // верхняя граница защищает от «бомб» в испорченных данных
	}
	if salt, err = base64.RawStdEncoding.DecodeString(parts[4]); err != nil || len(salt) == 0 {
		return p, nil, nil, ErrMalformedHash
	}
	if key, err = base64.RawStdEncoding.DecodeString(parts[5]); err != nil || len(key) == 0 {
		return p, nil, nil, ErrMalformedHash
	}

	return p, salt, key, nil
}
