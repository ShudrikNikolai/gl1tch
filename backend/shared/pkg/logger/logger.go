package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/reqctx"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	Level  string `yaml:"log_level"  env:"LOG_LEVEL"  envDefault:"info"`
	Format string `yaml:"log_format" env:"LOG_FORMAT" envDefault:"text"`
}

func MustSetupLogger(env string, cfg Config) *slog.Logger {
	log, err := newLogger(env, cfg)
	if err != nil {
		panic(err)
	}

	return log
}

func New(service, env string, cfg Config) *slog.Logger {
	return MustSetupLogger(env, cfg).With(slog.String("service", service))
}

func newLogger(env string, cfg Config) (*slog.Logger, error) {
	opts := &slog.HandlerOptions{Level: parseLevel(cfg.Level)}

	var h slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "json":
		h = slog.NewJSONHandler(os.Stdout, opts)
	default:
		h = slog.NewTextHandler(os.Stdout, opts)
	}

	// ctxHandler добавляет request_id и user_id из контекста, достаточно вызывать
	// log.InfoContext(ctx, ...), вручную их передавать не нужно!!!!!!
	return slog.New(ctxHandler{h}).With(slog.String("env", env)), nil
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type ctxHandler struct {
	slog.Handler
}

func (h ctxHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := reqctx.RequestID(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	if uid := reqctx.UserID(ctx); uid != "" {
		r.AddAttrs(slog.String("user_id", uid))
	}
	// OTel: trace_id и span_id
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		r.AddAttrs(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

func (h ctxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ctxHandler{h.Handler.WithAttrs(attrs)}
}

func (h ctxHandler) WithGroup(name string) slog.Handler {
	return ctxHandler{h.Handler.WithGroup(name)}
}
