package grpc_api

import (
	"fmt"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/protox"
	commonv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/common/v1"
	usersv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/users/v1"
	"github.com/ShudrikNikolai/gl1tch/backend/users/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func CreateUserRequestConverter(data *usersv1.CreateUserRequest) *model.User {
	return &model.User{
		UserID:   data.GetUserId(),
		Username: data.GetUsername(),
	}
}

func CreateUserResponseConverter(res *model.User) *usersv1.CreateUserResponse {
	return &usersv1.CreateUserResponse{User: userToProto(res)}
}

func MeResponseConverter(res *model.User) *usersv1.MeResponse {
	return &usersv1.MeResponse{User: userToProto(res)}
}

func GetUserResponseConverter(res *model.User) *usersv1.GetUserResponse {
	return &usersv1.GetUserResponse{User: userToProto(res)}
}

func ListUsersRequestConverter(data *usersv1.ListUsersRequest) *model.CursorRequest {
	return &model.CursorRequest{
		PageSize:  int32(data.GetCursor().GetPageSize()),
		PageToken: data.GetCursor().GetPageToken(),
	}
}

func ListUsersResponseConverter(users []model.User, cursor model.CursorResponse) *usersv1.ListUsersResponse {
	out := make([]*usersv1.User, 0, len(users))
	for i := range users {
		out = append(out, userToProto(&users[i]))
	}

	return &usersv1.ListUsersResponse{
		Users:  out,
		Cursor: cursorToProto(cursor),
	}
}

func userToProto(u *model.User) *usersv1.User {
	if u == nil {
		return nil
	}

	// TODO
	roles, _ := protox.RoleToProto("user")
	status, _ := protox.UserStatusToProto("active")
	gender, _ := protox.GenderToProto("male")
	//roles := make([]commonv1.Role, 1)
	//roles[0] =
	//for _, r := range u.Role {
	//	roles = append(roles, roleToProto(r))
	//}

	return &usersv1.User{
		UserId:      u.UserID,
		Username:    u.Username,
		Bio:         u.Bio,
		Role:        roles,
		Status:      status,
		Gender:      gender,
		CreatedAt:   timestampPtr(&u.CreatedAt),
		UpdatedAt:   timestampPtr(&u.UpdatedAt),
		DeletedAt:   timestampPtr(u.DeletedAt),
		BannedUntil: timestampPtr(u.BannedUntil),
	}
}

func cursorToProto(c model.CursorResponse) *commonv1.CursorPaginationResponse {
	return &commonv1.CursorPaginationResponse{
		NextPageToken: fmt.Sprintf("%d", c.NextPageToken),
		HasMore:       c.HasMore,
	}
}

func CursorRequestConverter(data *usersv1.ListUsersRequest) *model.CursorRequest {
	return &model.CursorRequest{
		PageSize:  int32(data.GetCursor().GetPageSize()),
		PageToken: data.GetCursor().GetPageToken(),
	}
}

func CursorResponseConverter(res *model.CursorResponse) *commonv1.CursorPaginationResponse {
	return &commonv1.CursorPaginationResponse{
		NextPageToken: fmt.Sprintf("%d", res.NextPageToken),
		HasMore:       res.HasMore,
	}
}

func timestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}

	return timestamppb.New(t)
}

func timestampPtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}

	return timestamppb.New(*t)
}

func timestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}

	return ts.AsTime()
}

func timestampPtrToTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}

	t := ts.AsTime()
	return &t
}
