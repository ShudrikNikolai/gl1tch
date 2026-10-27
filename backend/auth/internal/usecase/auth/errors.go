package auth

import (
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/models"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

var (
	ErrUserBanned   = apperr.PermissionDenied("USER_BANNED", "user is banned")
	ErrUserInactive = apperr.Unauthenticated("USER_INACTIVE", "user is not active")
)

// checkStatus пускает только активных пользователей.
func checkStatus(u models.User) error {
	switch u.Status {
	case models.UserStatusActive:
		return nil
	case models.UserStatusBanned:
		return ErrUserBanned
	default:
		return ErrUserInactive
	}
}
