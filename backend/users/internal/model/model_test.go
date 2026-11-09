package model

import (
	"testing"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/enum"
)

func TestRolesScan(t *testing.T) {
	var roles Roles
	if err := roles.Scan(`{user,moderator}`); err != nil {
		t.Fatal(err)
	}
	if len(roles) != 2 || roles[0] != enum.RoleUser || roles[1] != enum.RoleModerator {
		t.Fatalf("unexpected roles: %v", roles)
	}
}

func TestRolesScanEmpty(t *testing.T) {
	roles := Roles{enum.RoleUser}
	if err := roles.Scan(`{}`); err != nil {
		t.Fatal(err)
	}
	if len(roles) != 0 {
		t.Fatalf("expected empty roles, got %v", roles)
	}
}
