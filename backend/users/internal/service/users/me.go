package users

import (
	"context"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/model"
)

const rollbackTimeout = 5 * time.Second

func (s Service) Me(ctx context.Context, userID string) (model.User, error) {
	return s.users.Me(ctx, userID)
}
