package auth

import (
	"context"
	"log/slog"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/models"
)

// Refresh обменивает refresh-токен на новую пару. Роли и статус берутся из users заново,
// поэтому бан и смена роли вступают в силу не позже чем через время жизни access-токена.
func (uc *UseCase) Refresh(ctx context.Context, refreshToken string) (*models.RefreshResponse, error) {
	userID, next, err := uc.refresh.Rotate(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := uc.users.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := checkStatus(user); err != nil {
		// заблокированному пользователю новую цепочку токенов не оставляем
		if rerr := uc.refresh.RevokeAll(ctx, userID); rerr != nil {
			uc.log.ErrorContext(ctx, "refresh: revoke sessions failed",
				slog.String("user_id", userID), slog.Any("error", rerr))
		}

		return nil, err
	}

	tokens, err := uc.pair(user, next)
	if err != nil {
		return nil, err
	}

	return &models.RefreshResponse{Tokens: tokens}, nil
}
