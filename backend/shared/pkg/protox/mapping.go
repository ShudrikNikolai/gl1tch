package protox

import (
	"fmt"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/enum"
	commonv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/common/v1"
)

var roleFromProto = map[commonv1.Role]enum.Role{
	commonv1.Role_ROLE_UNSPECIFIED: enum.RoleUnspecified,
	commonv1.Role_ROLE_USER:        enum.RoleUser,
	commonv1.Role_ROLE_MODERATOR:   enum.RoleModerator,
	commonv1.Role_ROLE_ADMIN:       enum.RoleAdmin,
}

var roleToProto = map[enum.Role]commonv1.Role{
	enum.RoleUnspecified: commonv1.Role_ROLE_UNSPECIFIED,
	enum.RoleUser:        commonv1.Role_ROLE_USER,
	enum.RoleModerator:   commonv1.Role_ROLE_MODERATOR,
	enum.RoleAdmin:       commonv1.Role_ROLE_ADMIN,
}

func RoleFromProto(p commonv1.Role) (enum.Role, error) {
	m, ok := roleFromProto[p]
	if !ok {
		return enum.RoleUnspecified, fmt.Errorf("protox.RoleFromProto: unknown %d", p)
	}
	return m, nil
}

func RoleToProto(m enum.Role) (commonv1.Role, error) {
	p, ok := roleToProto[m]
	if !ok {
		return commonv1.Role_ROLE_UNSPECIFIED, fmt.Errorf("protox.RoleToProto: unknown %q", m)
	}
	return p, nil
}

var userStatusFromProto = map[commonv1.UserStatus]enum.UserStatus{
	commonv1.UserStatus_USER_STATUS_UNSPECIFIED: enum.UserStatusUnspecified,
	commonv1.UserStatus_USER_STATUS_ACTIVE:      enum.UserStatusActive,
	commonv1.UserStatus_USER_STATUS_DELETED:     enum.UserStatusDeleted,
	commonv1.UserStatus_USER_STATUS_BANNED:      enum.UserStatusBanned,
	commonv1.UserStatus_USER_STATUS_PENDING:     enum.UserStatusPending,
}

var userStatusToProto = map[enum.UserStatus]commonv1.UserStatus{
	enum.UserStatusUnspecified: commonv1.UserStatus_USER_STATUS_UNSPECIFIED,
	enum.UserStatusActive:      commonv1.UserStatus_USER_STATUS_ACTIVE,
	enum.UserStatusDeleted:     commonv1.UserStatus_USER_STATUS_DELETED,
	enum.UserStatusBanned:      commonv1.UserStatus_USER_STATUS_BANNED,
	enum.UserStatusPending:     commonv1.UserStatus_USER_STATUS_PENDING,
}

func UserStatusFromProto(p commonv1.UserStatus) (enum.UserStatus, error) {
	m, ok := userStatusFromProto[p]
	if !ok {
		return enum.UserStatusUnspecified, fmt.Errorf("protox.UserStatusFromProto: unknown %d", p)
	}
	return m, nil
}

func UserStatusToProto(m enum.UserStatus) (commonv1.UserStatus, error) {
	p, ok := userStatusToProto[m]
	if !ok {
		return commonv1.UserStatus_USER_STATUS_UNSPECIFIED, fmt.Errorf("protox.UserStatusToProto: unknown")
	}
	return p, nil
}

var genderFromProto = map[commonv1.Gender]enum.Gender{
	commonv1.Gender_GENDER_UNSPECIFIED: enum.GenderUnspecified,
	commonv1.Gender_GENDER_MALE:        enum.GenderMale,
	commonv1.Gender_GENDER_FEMALE:      enum.GenderFemale,
}

var genderToProto = map[enum.Gender]commonv1.Gender{
	enum.GenderUnspecified: commonv1.Gender_GENDER_UNSPECIFIED,
	enum.GenderMale:        commonv1.Gender_GENDER_MALE,
	enum.GenderFemale:      commonv1.Gender_GENDER_FEMALE,
}

func GenderFromProto(p commonv1.Gender) (enum.Gender, error) {
	m, ok := genderFromProto[p]
	if !ok {
		return enum.GenderUnspecified, fmt.Errorf("protox.GenderFromProto: unknown %d", p)
	}
	return m, nil
}

func GenderToProto(m enum.Gender) (commonv1.Gender, error) {
	p, ok := genderToProto[m]
	if !ok {
		return commonv1.Gender_GENDER_UNSPECIFIED, fmt.Errorf("protox.GenderToProto: unknown")
	}
	return p, nil
}

var orderFromProto = map[commonv1.Order]enum.Order{
	commonv1.Order_ORDER_UNSPECIFIED: enum.OrderUnspecified,
	commonv1.Order_ORDER_ASC:         enum.OrderAsc,
	commonv1.Order_ORDER_DESC:        enum.OrderDesc,
}

var orderToProto = map[enum.Order]commonv1.Order{
	enum.OrderUnspecified: commonv1.Order_ORDER_UNSPECIFIED,
	enum.OrderAsc:         commonv1.Order_ORDER_ASC,
	enum.OrderDesc:        commonv1.Order_ORDER_DESC,
}

func OrderFromProto(p commonv1.Order) (enum.Order, error) {
	m, ok := orderFromProto[p]
	if !ok {
		return enum.OrderUnspecified, fmt.Errorf("protox.OrderFromProto: unknown %d", p)
	}
	return m, nil
}

func OrderToProto(m enum.Order) (commonv1.Order, error) {
	p, ok := orderToProto[m]
	if !ok {
		return commonv1.Order_ORDER_UNSPECIFIED, fmt.Errorf("protox.OrderToProto: unknown")
	}

	return p, nil
}
