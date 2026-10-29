package models

import (
	"slices"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/enum"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/reqctx"
)

// Actor тот, кто вызывает usecase. Транспорт собирает его из заголовков x-user-id и x-roles.
type Actor struct {
	UserID string
	Roles  []enum.Role
}

func (a Actor) Authenticated() bool {
	return a.UserID != ""
}

// IsModerator: модераторские права есть у moder и admin.
func (a Actor) IsModerator() bool {
	return slices.Contains(a.Roles, reqctx.RoleModerator) || slices.Contains(a.Roles, reqctx.RoleAdmin)
}

// Viewer нужен репозиториям тем для фильтра видимости.
func (a Actor) Viewer() threads.Viewer {
	return threads.Viewer{UserID: a.UserID, Moderator: a.IsModerator()}
}

type UserStatus int

const (
	UserStatusUnknown UserStatus = iota
	UserStatusActive
	UserStatusBanned
	UserStatusDeleted
	UserStatusPending
)

// User то, что форум знает о профиле из сервиса users.
type User struct {
	ID       string
	Username string
	Status   UserStatus
}
