package users

import (
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/models"
	commonv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/common/v1"
	usersv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/users/v1"
)

func toUser(u *usersv1.User) models.User {
	return models.User{
		ID:       u.GetUserId(),
		Username: u.GetUsername(),
		Role:     toRole(u.GetRole()),
		Status:   toStatus(u.GetStatus()),
	}
}

// TODO
func toRole(r commonv1.Role) string {
	switch r {
	//case commonv1.Role_ROLE_ADMIN:
	//	return reqctx.RoleAdmin
	//case commonv1.Role_ROLE_MODERATOR:
	//	return reqctx.RoleModerator
	default:
		return "user"
	}
}

func toStatus(s commonv1.UserStatus) models.UserStatus {
	switch s {
	case commonv1.UserStatus_USER_STATUS_ACTIVE:
		return models.UserStatusActive
	case commonv1.UserStatus_USER_STATUS_BANNED:
		return models.UserStatusBanned
	case commonv1.UserStatus_USER_STATUS_DELETED:
		return models.UserStatusDeleted
	case commonv1.UserStatus_USER_STATUS_PENDING:
		return models.UserStatusPending
	default:
		return models.UserStatusUnknown
	}
}
