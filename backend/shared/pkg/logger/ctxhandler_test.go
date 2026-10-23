package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/reqctx"
)

func TestCtxHandlerAddsRequestFields(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(ctxHandler{slog.NewTextHandler(&buf, nil)}).With("k", "v")

	ctx := reqctx.WithUser(reqctx.WithRequestID(context.Background(), "req-1"), "user-7", nil)
	log.InfoContext(ctx, "hello")

	out := buf.String()
	for _, want := range []string{"request_id=req-1", "user_id=user-7", "k=v", "hello"} {
		if !strings.Contains(out, want) {
			t.Fatalf("log line %q misses %q", out, want)
		}
	}
}

func TestCtxHandlerWithoutContextData(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(ctxHandler{slog.NewTextHandler(&buf, nil)}).WithGroup("g")
	log.Info("plain")
	if strings.Contains(buf.String(), "request_id") {
		t.Fatalf("unexpected request_id in %q", buf.String())
	}
}

func TestParseLevelCaseInsensitive(t *testing.T) {
	if parseLevel("DEBUG") != slog.LevelDebug || parseLevel("Warn") != slog.LevelWarn || parseLevel("nope") != slog.LevelInfo {
		t.Fatal("level parsing is wrong")
	}
}
