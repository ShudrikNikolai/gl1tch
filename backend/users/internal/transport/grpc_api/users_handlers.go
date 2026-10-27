package grpc_api

import (
	"context"
	"errors"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/reqctx"
	usersv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/users/v1"
	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (u UsersServer) CreateUser(ctx context.Context, request *usersv1.CreateUserRequest) (*usersv1.CreateUserResponse, error) {
	data := CreateUserRequestConverter(request)
	println("raw: ", data)
	if err := u.userSrv.Create(ctx, data.UserID, data.Username); err != nil {
		if errors.Is(err, model.ErrUsernameTaken) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	created, err := u.userSrv.Get(ctx, data.UserID)
	println("created: ", data.UserID, err)

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	println("created: ", created.UserID, created.Username)

	return CreateUserResponseConverter(&created), nil
}

func (u UsersServer) ListUsers(ctx context.Context, request *usersv1.ListUsersRequest) (*usersv1.ListUsersResponse, error) {
	cursor := ListUsersRequestConverter(request)

	users, next, err := u.userSrv.List(ctx, cursor)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ListUsersResponseConverter(users, next), nil
}

func (u UsersServer) Me(ctx context.Context, request *usersv1.MeRequest) (*usersv1.MeResponse, error) {
	userID := reqctx.UserID(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	user, err := u.userSrv.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return MeResponseConverter(&user), nil
}

func (u UsersServer) GetUser(ctx context.Context, request *usersv1.GetUserRequest) (*usersv1.GetUserResponse, error) {
	user, err := u.userSrv.Get(ctx, request.GetUserId())
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return GetUserResponseConverter(&user), nil
}
