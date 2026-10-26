package users

import (
	"log/slog"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/repository"
	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/service"
)

var _ service.UsersService = (*Service)(nil)

type Deps struct {
	UsersRepo repository.Repository
	Log       *slog.Logger
}

type Service struct {
	users repository.Repository
	log   *slog.Logger
	now   func() time.Time
}

func New(d Deps) *Service {
	return &Service{
		users: d.UsersRepo,
		log:   d.Log,
		now:   time.Now,
	}
}
