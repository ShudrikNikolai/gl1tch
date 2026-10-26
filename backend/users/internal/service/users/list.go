package users

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/model"
)

func (s Service) List(ctx context.Context, cursor *model.CursorRequest) ([]model.User, model.CursorResponse, error) {
	return s.users.List(ctx, cursor)
}
