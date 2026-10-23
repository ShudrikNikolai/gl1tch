// Package config загружает настройки сервиса из переменных окружения
// в контейнерах: environment из *.yaml
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/logger"
)

// Base общие для всех сервисов настройки. Встраивайте в конфиг сервиса:
//
//	type Config struct {
//	    config.Base
//	    DatabaseURL string `env:"DATABASE_URL,required"`
//	}
type Base struct {
	Env          string `env:"ENV"        envDefault:"dev"`
	LogLevel     string `env:"LOG_LEVEL"  envDefault:"info"`
	LogFormat    string `env:"LOG_FORMAT" envDefault:"text"`
	HTTPPort     int    `env:"HTTP_PORT"  envDefault:"8080"`
	GRPCPort     int    `env:"GRPC_PORT"  envDefault:"9090"`
	OTLPEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
}

// Logger возвращает настройки логгера из Base.
func (b Base) Logger() logger.Config {
	return logger.Config{Level: b.LogLevel, Format: b.LogFormat}
}

// Load разбирает окружение в структуру T (теги env, required, envDefault).
func Load[T any]() (T, error) {
	return env.ParseAs[T]()
}

// MustLoad то же, что Load, но паникует при ошибке (для main).
func MustLoad[T any]() T {
	cfg, err := Load[T]()
	fmt.Println(&cfg, "\n", "\n", err)
	if err != nil {
		panic(err)
	}

	return cfg
}
