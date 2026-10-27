package models

type UserStatus int

const (
	UserStatusUnknown UserStatus = iota
	UserStatusActive
	UserStatusBanned
	UserStatusDeleted
	UserStatusPending
)

// User то, что auth знает о профиле из сервиса users.
type User struct {
	ID       string
	Username string
	Role     string // user | moder | admin (константы в shared/pkg/reqctx)
	Status   UserStatus
}
