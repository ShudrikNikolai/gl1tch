package auth

import (
	"context"
	"log/slog"
	"time"

	validation "github.com/ShudrikNikolai/gl1tch/backend/auth/internal/infrastructure/validate"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/models"
)

const rollbackTimeout = 5 * time.Second

// Register: учётка в auth -> профиль в users -> токены.
// Если профиль создать не удалось, учётка откатывается: иначе логин занят, а войти нельзя.
func (uc *UseCase) Register(ctx context.Context, req *models.BaseAuthRequest) (*models.BaseAuthResponse, error) {
	username, err := validation.Registration(req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	cred, err := uc.creds.Create(ctx, username, req.Password)
	if err != nil {
		return nil, err
	}

	user, err := uc.users.CreateUser(ctx, cred.UserID, cred.Username)
	if err != nil {
		uc.rollback(ctx, cred.UserID)

		return nil, err
	}

	tokens, err := uc.issueTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &models.BaseAuthResponse{UserId: user.ID, Tokens: tokens}, nil
}

// rollback выполняется даже если клиент уже отменил запрос.
func (uc *UseCase) rollback(ctx context.Context, userID string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), rollbackTimeout)
	defer cancel()

	if err := uc.creds.Delete(ctx, userID); err != nil {
		uc.log.ErrorContext(ctx, "register: rollback credentials failed (orphan account)",
			slog.String("user_id", userID), slog.Any("error", err))
	}
}
