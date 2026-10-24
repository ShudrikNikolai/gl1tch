package redis

import (
	"context"
	"log/slog"
	"time"

	redigo "github.com/gomodule/redigo/redis"
)

// Config настройки подключения (читаются из окружения через shared/pkg/config).
type Config struct {
	Addr              string        `env:"REDIS_ADDR,required"`
	Password          string        `env:"REDIS_PASSWORD"`
	DB                int           `env:"REDIS_DB"`
	MaxIdle           int           `env:"REDIS_MAX_IDLE"           envDefault:"10"`
	MaxActive         int           `env:"REDIS_MAX_ACTIVE"         envDefault:"50"`
	IdleTimeout       time.Duration `env:"REDIS_IDLE_TIMEOUT"       envDefault:"5m"`
	DialTimeout       time.Duration `env:"REDIS_DIAL_TIMEOUT"       envDefault:"2s"`
	ReadTimeout       time.Duration `env:"REDIS_READ_TIMEOUT"       envDefault:"2s"`
	WriteTimeout      time.Duration `env:"REDIS_WRITE_TIMEOUT"      envDefault:"2s"`
	ConnectionTimeout time.Duration `env:"REDIS_CONNECTION_TIMEOUT" envDefault:"2s"` // ожидание свободного соединения в пуле
}

// New создаёт клиент с пулом и поддержкой Subscribe.
func New(cfg Config, log *slog.Logger) *Client {
	baseOpts := func() []redigo.DialOption {
		opts := []redigo.DialOption{
			redigo.DialConnectTimeout(cfg.DialTimeout),
			redigo.DialReadTimeout(cfg.ReadTimeout),
			redigo.DialWriteTimeout(cfg.WriteTimeout),
			redigo.DialDatabase(cfg.DB),
		}
		if cfg.Password != "" {
			opts = append(opts, redigo.DialPassword(cfg.Password))
		}

		return opts
	}

	log.Info("New redis", "baseOpts", baseOpts)

	pool := &redigo.Pool{
		MaxIdle:     cfg.MaxIdle,
		MaxActive:   cfg.MaxActive,
		IdleTimeout: cfg.IdleTimeout,
		Wait:        true, // при исчерпании пула ждать, а не падать
		Dial: func() (redigo.Conn, error) {
			return redigo.Dial("tcp", cfg.Addr, baseOpts()...)
		},
		TestOnBorrow: func(c redigo.Conn, lastUsed time.Time) error {
			if time.Since(lastUsed) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")

			return err
		},
	}
	log.Info("New redis pool", "pool", pool)

	// Соединение подписки живёт долго: read timeout больше интервала keepalive-PING (см. pubsub.go).
	subDial := func(ctx context.Context) (redigo.Conn, error) {
		opts := append(baseOpts(), redigo.DialReadTimeout(subReadTimeout))
		return redigo.DialContext(ctx, "tcp", cfg.Addr, opts...)
	}
	log.Info("New redis pool after")

	return NewClient(pool, log, cfg.ConnectionTimeout, WithSubscriber(subDial))
}
