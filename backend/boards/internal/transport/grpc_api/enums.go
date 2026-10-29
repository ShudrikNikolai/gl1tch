package grpc_api

import (
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	boardsv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/boards/v1"
)

// unknown подставляется вместо значения enum, которого сервер не знает (клиент новее сервера).
// Такое значение не проходит валидацию в usecase и превращается в InvalidArgument, а не молча в «не задано».
const unknown = "unknown"

func visibilityFromProto(v boardsv1.Visibility) domain.Visibility {
	switch v {
	case boardsv1.Visibility_VISIBILITY_UNSPECIFIED:
		return ""
	case boardsv1.Visibility_VISIBILITY_PUBLIC:
		return domain.VisibilityPublic
	case boardsv1.Visibility_VISIBILITY_AUTHOR:
		return domain.VisibilityAuthor
	case boardsv1.Visibility_VISIBILITY_MODERATOR:
		return domain.VisibilityModerator
	default:
		return unknown
	}
}

func visibilityToProto(v domain.Visibility) boardsv1.Visibility {
	switch v {
	case domain.VisibilityPublic:
		return boardsv1.Visibility_VISIBILITY_PUBLIC
	case domain.VisibilityAuthor:
		return boardsv1.Visibility_VISIBILITY_AUTHOR
	case domain.VisibilityModerator:
		return boardsv1.Visibility_VISIBILITY_MODERATOR
	default:
		return boardsv1.Visibility_VISIBILITY_UNSPECIFIED
	}
}

func categoryFromProto(c boardsv1.ThreadCategory) threads.Category {
	switch c {
	case boardsv1.ThreadCategory_THREAD_CATEGORY_UNSPECIFIED:
		return ""
	case boardsv1.ThreadCategory_THREAD_CATEGORY_GENERAL:
		return threads.CategoryGeneral
	case boardsv1.ThreadCategory_THREAD_CATEGORY_FAQ:
		return threads.CategoryFAQ
	case boardsv1.ThreadCategory_THREAD_CATEGORY_FLOOD:
		return threads.CategoryFlood
	default:
		return unknown
	}
}

func categoryToProto(c threads.Category) boardsv1.ThreadCategory {
	switch c {
	case threads.CategoryGeneral:
		return boardsv1.ThreadCategory_THREAD_CATEGORY_GENERAL
	case threads.CategoryFAQ:
		return boardsv1.ThreadCategory_THREAD_CATEGORY_FAQ
	case threads.CategoryFlood:
		return boardsv1.ThreadCategory_THREAD_CATEGORY_FLOOD
	default:
		return boardsv1.ThreadCategory_THREAD_CATEGORY_UNSPECIFIED
	}
}

func statusFromProto(s boardsv1.ThreadStatus) threads.Status {
	switch s {
	case boardsv1.ThreadStatus_THREAD_STATUS_UNSPECIFIED:
		return ""
	case boardsv1.ThreadStatus_THREAD_STATUS_OPEN:
		return threads.StatusOpen
	case boardsv1.ThreadStatus_THREAD_STATUS_LOCKED:
		return threads.StatusLocked
	case boardsv1.ThreadStatus_THREAD_STATUS_ARCHIVED:
		return threads.StatusArchived
	default:
		return unknown
	}
}

func statusToProto(s threads.Status) boardsv1.ThreadStatus {
	switch s {
	case threads.StatusOpen:
		return boardsv1.ThreadStatus_THREAD_STATUS_OPEN
	case threads.StatusLocked:
		return boardsv1.ThreadStatus_THREAD_STATUS_LOCKED
	case threads.StatusArchived:
		return boardsv1.ThreadStatus_THREAD_STATUS_ARCHIVED
	default:
		return boardsv1.ThreadStatus_THREAD_STATUS_UNSPECIFIED
	}
}

func orderFromProto(o boardsv1.OrderByThread) threads.Order {
	switch o {
	case boardsv1.OrderByThread_ORDER_BY_THREAD_UNSPECIFIED:
		return ""
	case boardsv1.OrderByThread_ORDER_BY_THREAD_CREATED_AT:
		return threads.OrderCreated
	case boardsv1.OrderByThread_ORDER_BY_THREAD_LAST_POST_AT:
		return threads.OrderLastPost
	default:
		return unknown
	}
}
