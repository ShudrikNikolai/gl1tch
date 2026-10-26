package grpc_api

import (
	"testing"
	"time"
)

func TestCookieFormat(t *testing.T) {
	c := CookieConfig{Name: "rt", Path: "/api/auth", Secure: true, MaxAge: 720 * time.Hour}

	if got, want := c.set("tok"), "rt=tok; HttpOnly; SameSite=Strict; Path=/api/auth; Max-Age=2592000; Secure"; got != want {
		t.Fatalf("set:\n got %q\nwant %q", got, want)
	}
	if got, want := c.clear(), "rt=; HttpOnly; SameSite=Strict; Path=/api/auth; Max-Age=0; Secure"; got != want {
		t.Fatalf("clear:\n got %q\nwant %q", got, want)
	}

	c.Secure = false
	if got, want := c.set("tok"), "rt=tok; HttpOnly; SameSite=Strict; Path=/api/auth; Max-Age=2592000"; got != want {
		t.Fatalf("insecure set:\n got %q\nwant %q", got, want)
	}
}
