package auth

import (
	"context"

	validation "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/validate"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/models"
)

func (uc *UseCase) Login(ctx context.Context, req *models.BaseAuthRequest) (*models.BaseAuthResponse, error) {
	username, err := validation.Login(req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	cred, err := uc.creds.Authenticate(ctx, username, req.Password)
	if err != nil {
		return nil, err
	}

	user, err := uc.users.GetUser(ctx, cred.UserID)
	if err != nil {
		return nil, err
	}
	if err = checkStatus(user); err != nil {
		return nil, err
	}

	tokens, err := uc.issueTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &models.BaseAuthResponse{UserId: user.ID, Tokens: tokens}, nil
}
