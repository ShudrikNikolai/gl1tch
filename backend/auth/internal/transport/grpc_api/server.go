package grpc_api

import (
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/auth"
	authv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/auth/v1"
	"google.golang.org/grpc"
)

type AuthGrpcServer struct {
	authv1.UnimplementedAuthServiceServer
	authcase *auth.UseCase
	//	grpctosrv
}

func RegisterAuthGrpcServer(s grpc.ServiceRegistrar, authcase *auth.UseCase) {
	authv1.RegisterAuthServiceServer(s, &AuthGrpcServer{authcase: authcase})
}
