package grpc_api

import (
	"context"

	authv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/auth/v1"
	//_ "google.golang.org/grpc"// Установка куки
	//_ "google.golang.org/grpc/metadata"
)

func (ags *AuthGrpcServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	//reqToSrv := grpctosrv.RegisterConverter(req)
	//result := ags.authcase.Register(ctx, reqToSrv)
	//srvToReq := &authv1.RegisterResponse{}
	// Установка куки
	//cookie := "rt=" + refresh + "; HttpOnly; Secure; SameSite=Strict; Path=/api/auth; Max-Age=2592000"
	//_ = grpc.SetHeader(ctx, metadata.Pairs("set-cookie", cookie))
	return &authv1.RegisterResponse{}, nil
}

func (ags *AuthGrpcServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return &authv1.LoginResponse{}, nil
}

func (ags *AuthGrpcServer) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	return &authv1.LogoutResponse{}, nil
}

func (ags *AuthGrpcServer) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	return &authv1.RefreshResponse{}, nil
}
