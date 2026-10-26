package config

import (
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/config"
)

type Config struct {
	config.Base

	Database        Database
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

type Database struct {
	DatabaseURL string `env:"DATABASE_URL,required"`
}
