package reqctx

import (
	"context"
	"testing"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/enum"
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
	ctx := WithUser(context.Background(), "u1", []enum.Role{enum.RoleUser, enum.RoleModerator})
	if UserID(ctx) != "u1" {
		t.Fatal("user id lost")
	}
	if !HasRole(ctx, enum.RoleModerator) || !HasRole(ctx, enum.RoleUser) {
		t.Fatal("roles lost")
	}
	if HasRole(context.Background(), enum.RoleUser) {
		t.Fatal("anonymous must have no roles")
	}
}

func TestRolesAreCopied(t *testing.T) {
	in := []enum.Role{enum.RoleUser}
	ctx := WithUser(context.Background(), "u1", in)
	in[0] = enum.RoleModerator
	if !HasRole(ctx, enum.RoleUser) || HasRole(ctx, enum.RoleModerator) {
		t.Fatal("context must hold its own copy of roles")
	}
	out := Roles(ctx)
	out[0] = enum.RoleModerator
	if HasRole(ctx, enum.RoleModerator) {
		t.Fatal("Roles must return a copy")
	}
}
