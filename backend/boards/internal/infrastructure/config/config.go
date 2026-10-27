package config

import (
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/config"
)

type Config struct {
	config.Base

	Database         Database
	ShutdownTimeout  time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	ServicesGRPCAddr ServicesGRPCAddr

	// Профили из users кешируются: бан доходит до форума не позже чем через это время.
	UsersCacheTTL time.Duration `env:"USERS_CACHE_TTL" envDefault:"1m"`
}

type Database struct {
	DatabaseURL string `env:"DATABASE_URL,required"`
}

type ServicesGRPCAddr struct {
	UsersGRPCAddr string `env:"USERS_GRPC_ADDR,required"`
}
