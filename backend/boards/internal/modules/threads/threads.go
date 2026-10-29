// Package threads описывает темы: заголовок, статус, закрепление и сводку по последней активности.
package threads

import (
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

type Status string

const (
	StatusOpen     Status = "open"
	StatusLocked   Status = "locked" // только чтение
	StatusArchived Status = "archived"
)

func (s Status) Valid() bool {
	return s == StatusOpen || s == StatusLocked || s == StatusArchived
}

type Category string

const (
	CategoryGeneral Category = "general"
	CategoryFAQ     Category = "faq"
	CategoryFlood   Category = "flood"
)

func (c Category) Valid() bool {
	return c == CategoryGeneral || c == CategoryFAQ || c == CategoryFlood
}

// Order порядок списка тем. Закреплённые темы всегда идут первыми.
type Order string

const (
	OrderLastPost Order = "last_post" // «бамп»: тема с новым ответом поднимается (по умолчанию)
	OrderCreated  Order = "created"
)

type Thread struct {
	ID      string
	BoardID string
	Title   string
	Slug    string

	Author domain.Author

	Status     Status
	Category   Category
	Visibility domain.Visibility
	Pinned     bool

	PostCount int // не удалённые посты, включая первый

	LastPostAuthor domain.Author
	LastPostAt     time.Time

	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Viewer тот, кто читает: нужен, чтобы скрывать темы с ограниченной видимостью.
type Viewer struct {
	UserID    string
	Moderator bool
}

// VisibleTo: public видят все, author автор и модераторы, moderator только модераторы.
func (t Thread) VisibleTo(v Viewer) bool {
	switch t.Visibility {
	case domain.VisibilityPublic:
		return true
	case domain.VisibilityAuthor:
		return v.Moderator || (v.UserID != "" && v.UserID == t.Author.UserID)
	default:
		return v.Moderator
	}
}

// AcceptsPosts: писать можно только в открытую тему.
func (t Thread) AcceptsPosts() bool {
	return t.Status == StatusOpen
}

// ListFilter пустые поля означают «без фильтра». BoardID обязателен.
type ListFilter struct {
	BoardID  string
	Category Category
	Status   Status
	AuthorID string
	Search   string // подстрока заголовка, без учёта регистра
	Order    Order
	Viewer   Viewer
}

// UpdateParams nil означает «не менять».
type UpdateParams struct {
	Title      *string
	Category   *Category
	Visibility *domain.Visibility
}

var (
	ErrNotFound  = apperr.NotFound("THREAD_NOT_FOUND", "thread not found")
	ErrSlugTaken = apperr.AlreadyExists("THREAD_SLUG_TAKEN", "thread slug is already taken")
)
