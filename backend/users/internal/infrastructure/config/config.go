package config

import (
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/config"
)

type Config struct {
	config.Base

	Database        Database
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`

	ServicesGRPCAddr ServicesGRPCAddr

	// Вычистка зависших регистраций (учётка есть, профиля в users нет).
	//PendingMaxAge time.Duration `env:"PENDING_MAX_AGE" envDefault:"2m"`
	//PurgeEvery    time.Duration `env:"PURGE_EVERY"     envDefault:"1m"`
}

type Database struct {
	DatabaseURL string `env:"DATABASE_URL,required"`
	//DatabaseName string `env:"DATABASE_NAME" envDefault:"gl1tch-auth"`
	//DatabaseUser string `env:"DATABASE_USER" envDefault:"gl1tch-auth"`
	//DatabasePass string `env:"DATABASE_PASS" envDefault:"gl1tch-auth"`
}

type ServicesGRPCAddr struct {
	// UsersGRPCAddr string `env:"USERS_GRPC_ADDR,required"`
}
