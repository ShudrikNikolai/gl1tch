package grpc_api

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/reqctx"
)

// actorFrom собирает вызывающего из контекста: интерцептор grpcx кладёт туда x-user-id и x-roles, которые выставил Envoy.
func actorFrom(ctx context.Context) models.Actor {
	// TODO переделать потом и привести к одному виду, всю эту хуйню

	return models.Actor{UserID: reqctx.UserID(ctx), Roles: reqctx.Roles(ctx)}
}
