package model

import (
	"errors"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/enum"
)

type User struct {
	UserID   string `db:"user_id"   json:"user_id"`
	Username string `db:"username"  json:"username"`
	Bio      string `db:"bio"       json:"bio"`

	Gender     enum.Gender     `db:"gender"      json:"gender"`
	Role       []string        `db:"roles"       json:"roles"` //| []enum.Role
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
	NextPageToken int32
	HasMore       bool
}

var (
	ErrNotFound = errors.New("users: not found")

	ErrUsernameTaken = apperr.AlreadyExists("USERNAME_TAKEN", "username is already taken")
)
