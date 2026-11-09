package config

import (
	"strings"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/config"
)

type Config struct {
	config.Base

	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`

	KafkaBrokers string `env:"KAFKA_BROKERS" envDefault:"kafka:9092"`
	KafkaTopic   string `env:"KAFKA_TOPIC"   envDefault:"posts.events"`
	KafkaGroup   string `env:"KAFKA_GROUP"   envDefault:"notifications"`

	JWKSURL        string        `env:"JWKS_URL"            envDefault:"http://auth:8080/.well-known/jwks.json"`
	JWTIssuer      string        `env:"JWT_ISSUER"          envDefault:"gl1tch-auth"`
	JWKSRefresh    time.Duration `env:"JWKS_REFRESH_INTERVAL" envDefault:"10m"`
	JWKSMinRefresh time.Duration `env:"JWKS_MIN_REFRESH_GAP"  envDefault:"10s"`
	BoardsGRPCAddr string        `env:"BOARDS_GRPC_ADDR"    envDefault:"boards:9090"`
}

func (c Config) Brokers() []string {
	parts := strings.Split(c.KafkaBrokers, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}

	return out
}
