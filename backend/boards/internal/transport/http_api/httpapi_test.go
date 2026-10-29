package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type pinger struct{ err error }

func (p pinger) Ping(context.Context) error { return p.err }

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

	return rec
}

func TestHealthz(t *testing.T) {
	if rec := get(New(nil), "/healthz"); rec.Code != http.StatusOK {
		t.Fatalf("code = %d", rec.Code)
	}
}

func TestReadyz(t *testing.T) {
	ok := New(map[string]Pinger{"postgres": pinger{}})
	if rec := get(ok, "/readyz"); rec.Code != http.StatusOK {
		t.Fatalf("healthy: code = %d", rec.Code)
	}

	down := New(map[string]Pinger{"postgres": pinger{err: errors.New("dial tcp 10.0.0.5:5432: password=secret")}})
	rec := get(down, "/readyz")
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), `"postgres"`) {
		t.Fatalf("unhealthy: code = %d body = %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret") || strings.Contains(rec.Body.String(), "10.0.0.5") {
		t.Fatalf("the response must not leak error details: %s", rec.Body.String())
	}
}
