package usecase

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/models"
)

// AuthUseCase то, что транспорт знает о бизнес-логике. Ошибки возвращаются как apperr.
type AuthUseCase interface {
	Register(ctx context.Context, req *models.BaseAuthRequest) (*models.BaseAuthResponse, error)
	Login(ctx context.Context, req *models.BaseAuthRequest) (*models.BaseAuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*models.RefreshResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}
