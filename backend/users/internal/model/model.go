package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/enum"
)

// Roles — text[] user_role[] из Postgres в доменном виде.
type Roles []enum.Role

func (r *Roles) Scan(src any) error {
	var raw string
	switch v := src.(type) {
	case string:
		raw = v
	case []byte:
		raw = string(v)
	case nil:
		*r = nil
		return nil
	default:
		return fmt.Errorf("users: roles: unsupported scan type %T", src)
	}

	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		*r = nil
		return nil
	}
	raw = strings.TrimPrefix(strings.TrimSuffix(raw, "}"), "{")

	parts := strings.Split(raw, ",")
	roles := make(Roles, 0, len(parts))
	for _, p := range parts {
		p = strings.Trim(strings.TrimSpace(p), `"`)
		if p != "" {
			roles = append(roles, enum.Role(p))
		}
	}
	*r = roles

	return nil
}

type User struct {
	UserID   string `db:"user_id"   json:"user_id"`
	Username string `db:"username"  json:"username"`
	Bio      string `db:"bio"       json:"bio"`

	Gender     enum.Gender     `db:"gender"      json:"gender"`
	Role       Roles           `db:"roles"       json:"roles"`
	UserStatus enum.UserStatus `db:"user_status" json:"user_status"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	DeletedAt   *time.Time `db:"deleted_at"   json:"deleted_at,omitempty"`
	BannedUntil *time.Time `db:"banned_until" json:"banned_until,omitempty"`
}

type CursorRequest struct {
	PageSize  int32
	PageToken string
}

type CursorResponse struct {
	NextPageToken string
	HasMore       bool
}

var (
	ErrNotFound = apperr.NotFound("USER_NOT_FOUND", "user not found")

	ErrUsernameTaken = apperr.AlreadyExists("USERNAME_TAKEN", "username is already taken")
)
