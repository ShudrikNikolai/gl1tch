package grpc_api

import (
	"time"

	modelsUseCase "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/models"
	authv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/auth/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func RegisterResponseConverter(res *modelsUseCase.BaseAuthResponse) *authv1.RegisterResponse {
	return &authv1.RegisterResponse{UserId: res.UserId, Tokens: tokenPairConverter(res.Tokens)}
}

func LoginResponseConverter(res *modelsUseCase.BaseAuthResponse) *authv1.LoginResponse {
	return &authv1.LoginResponse{UserId: res.UserId, Tokens: tokenPairConverter(res.Tokens)}
}

func RefreshResponseConverter(res *modelsUseCase.RefreshResponse) *authv1.RefreshResponse {
	return &authv1.RefreshResponse{Tokens: tokenPairConverter(res.Tokens)}
}

func tokenPairConverter(t modelsUseCase.TokenPair) *authv1.TokenPair {
	return &authv1.TokenPair{
		AccessToken:      t.AccessToken,
		RefreshToken:     t.RefreshToken,
		AccessExpiresAt:  timestamp(t.AccessExpiresAt),
		RefreshExpiresAt: timestamp(t.RefreshExpiresAt),
	}
}

func timestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}

	return timestamppb.New(*t)
}
