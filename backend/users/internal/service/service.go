package service

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/model"
)

type UsersService interface {
	Create(ctx context.Context, userID, username string) error
	Me(ctx context.Context, userID string) (model.User, error)
	List(ctx context.Context, cursor *model.CursorRequest) ([]model.User, model.CursorResponse, error)
	Get(ctx context.Context, userID string) (model.User, error)
}
