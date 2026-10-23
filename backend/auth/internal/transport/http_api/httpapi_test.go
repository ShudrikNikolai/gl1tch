package httpapi

import (
	"context"
	"encoding/json"
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

func TestJWKS(t *testing.T) {
	h := New(func() []byte { return []byte(`{"keys":[]}`) }, nil)
	rec := get(h, "/.well-known/jwks.json")
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "application/json" ||
		!strings.Contains(rec.Header().Get("Cache-Control"), "max-age") || rec.Body.String() != `{"keys":[]}` {
		t.Fatalf("%d %v %q", rec.Code, rec.Header(), rec.Body.String())
	}
}

func TestHealthz(t *testing.T) {
	if rec := get(New(func() []byte { return nil }, nil), "/healthz"); rec.Code != 200 {
		t.Fatalf("healthz: %d", rec.Code)
	}
}

func TestReadyz(t *testing.T) {
	ok := New(func() []byte { return nil }, map[string]Pinger{"postgres": pinger{}, "redis": pinger{}})
	if rec := get(ok, "/readyz"); rec.Code != 200 {
		t.Fatalf("all dependencies up: %d", rec.Code)
	}

	bad := New(func() []byte { return nil }, map[string]Pinger{
		"postgres": pinger{err: errors.New("connect to 10.0.0.5 password=secret")}, "redis": pinger{},
	})
	rec := get(bad, "/readyz")
	var body struct {
		Failed []string `json:"failed"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != 503 || len(body.Failed) != 1 || body.Failed[0] != "postgres" {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret") || strings.Contains(rec.Body.String(), "10.0.0.5") {
		t.Fatal("probe response must not leak error details")
	}
}

func TestOnlyGET(t *testing.T) {
	h := New(func() []byte { return nil }, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/healthz", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST must be rejected, got %d", rec.Code)
	}
}
