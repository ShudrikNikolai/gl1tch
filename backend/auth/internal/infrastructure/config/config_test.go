package config

import (
	"testing"
	"time"

	sharedconfig "github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/config"
)

func TestDefaultsAndRequired(t *testing.T) {
	if _, err := sharedconfig.Load[Config](); err == nil {
		t.Fatal("required variables must be enforced")
	}

	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("REDIS_ADDR", "127.0.0.1:6379")
	t.Setenv("JWT_PRIVATE_KEY_FILE", "/run/secrets/jwt_private_key")
	t.Setenv("USERS_GRPC_ADDR", "dns:///localhost:9090")
	cfg, err := sharedconfig.Load[Config]()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Security.AccessTTL != 10*time.Minute ||
		cfg.Security.RefreshTTL != 720*time.Hour ||
		cfg.Security.RefreshReuseGrace != 5*time.Second ||
		cfg.Security.JWTIssuer != "gl1tch-auth" ||
		cfg.Security.RevocationChannel != "auth.revocations" {
		t.Fatalf("defaults: %+v", cfg)
	}
	if cfg.HTTPPort != 8080 || cfg.GRPCPort != 9090 {
		t.Fatalf("base defaults: %d %d", cfg.HTTPPort, cfg.GRPCPort)
	}
	if cfg.Argon2.MemoryKiB != 19456 ||
		cfg.Argon2.Iterations != 2 ||
		cfg.Redis.MaxActive != 50 {
		t.Fatalf("nested defaults: %+v %+v", cfg.Argon2, cfg.Redis)
	}
}
