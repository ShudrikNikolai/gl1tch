package config

import "testing"

type testCfg struct {
	Base
	DatabaseURL string `env:"DATABASE_URL,required"`
}

func TestLoadDefaultsAndRequired(t *testing.T) {
	if _, err := Load[testCfg](); err == nil {
		t.Fatal("expected error for missing DATABASE_URL")
	}

	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("HTTP_PORT", "8082")
	cfg, err := Load[testCfg]()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPPort != 8082 || cfg.GRPCPort != 9090 || cfg.Env != "dev" || cfg.DatabaseURL != "postgres://x" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.Logger().Level != "info" || cfg.Logger().Format != "text" {
		t.Fatalf("unexpected logger config: %+v", cfg.Logger())
	}
}
