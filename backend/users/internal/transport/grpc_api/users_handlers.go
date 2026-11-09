package grpc_api

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/reqctx"
	usersv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/users/v1"
)

func (u UsersServer) CreateUser(ctx context.Context, request *usersv1.CreateUserRequest) (*usersv1.CreateUserResponse, error) {
	data := CreateUserRequestConverter(request)
	if err := u.userSrv.Create(ctx, data.UserID, data.Username); err != nil {
		return nil, err
	}

	created, err := u.userSrv.Get(ctx, data.UserID)
	if err != nil {
		return nil, err
	}

	return CreateUserResponseConverter(&created), nil
}

func (u UsersServer) ListUsers(ctx context.Context, request *usersv1.ListUsersRequest) (*usersv1.ListUsersResponse, error) {
	cursor := ListUsersRequestConverter(request)

	users, next, err := u.userSrv.List(ctx, cursor)
	if err != nil {
		return nil, err
	}

	return ListUsersResponseConverter(users, next), nil
}

func (u UsersServer) Me(ctx context.Context, _ *usersv1.MeRequest) (*usersv1.MeResponse, error) {
	userID := reqctx.UserID(ctx)
	if userID == "" {
		return nil, apperr.Unauthenticated("UNAUTHENTICATED", "authentication required")
	}

	user, err := u.userSrv.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	return MeResponseConverter(&user), nil
}

func (u UsersServer) GetUser(ctx context.Context, request *usersv1.GetUserRequest) (*usersv1.GetUserResponse, error) {
	user, err := u.userSrv.Get(ctx, request.GetUserId())
	if err != nil {
		return nil, err
	}

	return GetUserResponseConverter(&user), nil
}
