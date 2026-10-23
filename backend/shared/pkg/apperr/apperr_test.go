package apperr

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestDefaultCodeAndFrom(t *testing.T) {
	e := New(KindNotFound, "", "no such user")
	if e.Code != CodeNotFound {
		t.Fatalf("code %q", e.Code)
	}
	wrapped := fmt.Errorf("handler: %w", NotFound("USER_NOT_FOUND", "no such user"))
	got := From(wrapped)
	if got.Kind != KindNotFound || got.Code != "USER_NOT_FOUND" {
		t.Fatalf("unwrapped: %+v", got)
	}
}

func TestUnknownErrorsAreInternalAndHidden(t *testing.T) {
	cause := errors.New("pq: connection refused at 10.0.0.5")
	e := From(cause)
	if e.Kind != KindInternal || e.Code != CodeInternal || e.Message != "internal error" {
		t.Fatalf("%+v", e)
	}
	if !errors.Is(e, cause) {
		t.Fatal("cause must stay reachable for logs")
	}
	if From(nil) != nil {
		t.Fatal("From(nil) must be nil")
	}
}

func TestHTTPStatus(t *testing.T) {
	cases := map[Kind]int{
		KindInvalidArgument: http.StatusBadRequest, KindUnauthenticated: http.StatusUnauthorized,
		KindPermissionDenied: http.StatusForbidden, KindNotFound: http.StatusNotFound,
		KindAlreadyExists: http.StatusConflict, KindRateLimited: http.StatusTooManyRequests,
		KindUnavailable: http.StatusServiceUnavailable, KindInternal: http.StatusInternalServerError,
	}
	for k, want := range cases {
		if got := HTTPStatus(k); got != want {
			t.Errorf("kind %d: got %d want %d", k, got, want)
		}
	}
}

func TestWrapKeepsCause(t *testing.T) {
	cause := errors.New("db")
	e := Wrap(KindUnavailable, "", "try later", cause)
	if !errors.Is(e, cause) || e.Code != CodeUnavailable {
		t.Fatalf("%+v", e)
	}
}
