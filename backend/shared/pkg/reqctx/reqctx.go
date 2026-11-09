// Package reqctx хранит в context данные запроса, общие для HTTP и gRPC:
// request id и аутентифицированного пользователя
// имена заголовков и ключей метаданных
package reqctx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"slices"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/enum"
)

// Имена заголовков HTTP и ключей метаданных gRPC (gateway -> сервисы)
const (
	HeaderRequestID   = "x-request-id"
	HeaderUserID      = "x-user-id"
	HeaderRoles       = "x-roles"
	HeaderTransparent = "transparent"
)

const (
	RoleUser      = enum.RoleUser
	RoleModerator = enum.RoleModerator
	RoleAdmin     = enum.RoleAdmin
)

type ctxKey int

const (
	requestIDKey ctxKey = iota
	userIDKey
	rolesKey
)

func NewRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID возвращает request id или пустую строку
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// WithUser сохраняет id пользователя и его роли (копия среза)
func WithUser(ctx context.Context, userID string, roles []enum.Role) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	return context.WithValue(ctx, rolesKey, slices.Clone(roles))
}

// UserID возвращает id пользователя или пустую строку
func UserID(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey).(string)
	return id
}

// Roles возвращает копию ролей пользователя
func Roles(ctx context.Context) []enum.Role {
	roles, _ := ctx.Value(rolesKey).([]enum.Role)
	return slices.Clone(roles)
}

func HasRole(ctx context.Context, role enum.Role) bool {
	roles, _ := ctx.Value(rolesKey).([]enum.Role)
	return slices.Contains(roles, role)
}
