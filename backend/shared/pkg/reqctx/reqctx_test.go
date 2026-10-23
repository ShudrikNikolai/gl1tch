package reqctx

import (
	"context"
	"testing"
)

func TestRequestID(t *testing.T) {
	ctx := context.Background()
	if RequestID(ctx) != "" {
		t.Fatal("empty context must have no request id")
	}
	ctx = WithRequestID(ctx, "abc")
	if got := RequestID(ctx); got != "abc" {
		t.Fatalf("got %q", got)
	}
}

func TestUserAndRoles(t *testing.T) {
	ctx := WithUser(context.Background(), "u1", []string{RoleUser, RoleModerator})
	if UserID(ctx) != "u1" {
		t.Fatal("user id lost")
	}
	if !HasRole(ctx, RoleModerator) || !HasRole(ctx, RoleUser) {
		t.Fatal("roles lost")
	}
	if HasRole(context.Background(), RoleUser) {
		t.Fatal("anonymous must have no roles")
	}
}

func TestRolesAreCopied(t *testing.T) {
	in := []string{RoleUser}
	ctx := WithUser(context.Background(), "u1", in)
	in[0] = RoleModerator
	if !HasRole(ctx, RoleUser) || HasRole(ctx, RoleModerator) {
		t.Fatal("context must hold its own copy of roles")
	}
	out := Roles(ctx)
	out[0] = RoleModerator
	if HasRole(ctx, RoleModerator) {
		t.Fatal("Roles must return a copy")
	}
}
