// Package reqctx хранит в context данные запроса, общие для HTTP и gRPC:
// request id и аутентифицированного пользователя
// имена заголовков и ключей метаданных
package reqctx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"slices"
)

// Имена заголовков HTTP и ключей метаданных gRPC (gateway -> сервисы)
const (
	HeaderRequestID   = "x-request-id"
	HeaderUserID      = "x-user-id"
	HeaderRoles       = "x-roles"
	HeaderTraceparent = "traceparent"
)

// Роли пользователей
const (
	RoleUser      = "user"
	RoleModerator = "moder"
	RoleAdmin     = "admin"
)

type ctxKey int

const (
	requestIDKey ctxKey = iota
	userIDKey
	rolesKey
)

// NewRequestID генерирует идентификатор запроса (16 случайных байт в hex)
// //TODO заменить uuid
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
func WithUser(ctx context.Context, userID string, roles []string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	return context.WithValue(ctx, rolesKey, slices.Clone(roles))
}

// UserID возвращает id пользователя или пустую строку
func UserID(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey).(string)
	return id
}

// Roles возвращает копию ролей пользователя
func Roles(ctx context.Context) []string {
	roles, _ := ctx.Value(rolesKey).([]string)
	return slices.Clone(roles)
}

func HasRole(ctx context.Context, role string) bool {
	roles, _ := ctx.Value(rolesKey).([]string)
	return slices.Contains(roles, role)
}
