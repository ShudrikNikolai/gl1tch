package users

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/model"
)

func (s Service) Get(ctx context.Context, userID string) (model.User, error) {
	return s.users.Get(ctx, userID)
}
