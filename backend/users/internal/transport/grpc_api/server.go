package grpc_api

import (
	usersv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/users/v1"
	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/service"
	"google.golang.org/grpc"
)

type UsersServer struct {
	usersv1.UnimplementedUserServiceServer
	userSrv service.UsersService // TODO
}

func RegisterUsersGrpcServer(srv *grpc.Server, userSrv service.UsersService) {
	usersv1.RegisterUserServiceServer(srv, &UsersServer{
		userSrv: userSrv,
	})
}
