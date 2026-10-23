package usecase

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/models"
)

type AuthUseCase interface {
	//Login()
	//Logout()
	//Refresh()
	Register(ctx context.Context, data *models.BaseAuthRequest) *models.BaseAuthResponse
}
