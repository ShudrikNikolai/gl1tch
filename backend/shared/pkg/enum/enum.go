// enum: role, status, gender, order
// константы, todo потом расписать нормально
package enum

type Role string

const (
	RoleUnspecified Role = "unspecified"
	RoleUser        Role = "user"
	RoleModerator   Role = "moderator"
	RoleAdmin       Role = "admin"
)

type UserStatus string

const (
	UserStatusUnspecified UserStatus = "unspecified"
	UserStatusActive      UserStatus = "active"
	UserStatusBanned      UserStatus = "banned"
	UserStatusDeleted     UserStatus = "deleted"
	UserStatusPending     UserStatus = "pending"
)

type Gender string

const (
	GenderUnspecified Gender = "unspecified"
	GenderMale        Gender = "male"
	GenderFemale      Gender = "female"
)

type Order string

const (
	OrderUnspecified Order = "unspecified"
	OrderAsc         Order = "asc"
	OrderDesc        Order = "desc"
)
