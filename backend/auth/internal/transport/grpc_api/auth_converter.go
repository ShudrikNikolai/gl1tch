package grpc_api

import (
	modelsUseCase "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/models"
	authv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/auth/v1"
)

func LoginConverter(data *authv1.LoginRequest) *modelsUseCase.BaseAuthRequest {
	return &modelsUseCase.BaseAuthRequest{
		Username: data.GetUsername(),
		Password: data.GetPassword(),
	}
}

func RegisterConverter(data *authv1.RegisterRequest) *modelsUseCase.BaseAuthRequest {
	return &modelsUseCase.BaseAuthRequest{
		Username: data.GetUsername(),
		Password: data.GetPassword(),
	}
}
