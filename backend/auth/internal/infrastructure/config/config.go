package config

import (
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/cache/redis"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/config"
)

type Config struct {
	config.Base

	Database        Database
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	Redis           redis.Config

	Security         Security
	ServicesGRPCAddr ServicesGRPCAddr

	// Вычистка зависших регистраций (учётка есть, профиля в users нет).
	//PendingMaxAge time.Duration `env:"PENDING_MAX_AGE" envDefault:"2m"`
	//PurgeEvery    time.Duration `env:"PURGE_EVERY"     envDefault:"1m"`

	Argon2 Argon2
}

type Database struct {
	DatabaseURL string `env:"DATABASE_URL,required"`
	//DatabaseName string `env:"DATABASE_NAME" envDefault:"gl1tch-auth"`
	//DatabaseUser string `env:"DATABASE_USER" envDefault:"gl1tch-auth"`
	//DatabasePass string `env:"DATABASE_PASS" envDefault:"gl1tch-auth"`
}

type Security struct {
	JWTPrivateKeyFile string        `env:"JWT_PRIVATE_KEY_FILE,required"`
	JWTIssuer         string        `env:"JWT_ISSUER"         envDefault:"gl1tch-auth"`
	AccessTTL         time.Duration `env:"ACCESS_TOKEN_TTL"   envDefault:"10m"`
	RefreshTTL        time.Duration `env:"REFRESH_TOKEN_TTL"  envDefault:"720h"`

	RefreshReuseGrace time.Duration `env:"REFRESH_REUSE_GRACE" envDefault:"5s"`
	RevocationChannel string        `env:"REVOCATION_CHANNEL" envDefault:"auth.revocations"`
}

type Argon2 struct {
	MemoryKiB     uint32 `env:"ARGON2_MEMORY_KIB"     envDefault:"19456"`
	Iterations    uint32 `env:"ARGON2_ITERATIONS"     envDefault:"2"`
	Parallelism   uint8  `env:"ARGON2_PARALLELISM"    envDefault:"1"`
	MaxConcurrent int    `env:"ARGON2_MAX_CONCURRENT" envDefault:"0"` // 0 = число CPU
}

type ServicesGRPCAddr struct {
	UsersGRPCAddr string `env:"USERS_GRPC_ADDR,required"`
}
