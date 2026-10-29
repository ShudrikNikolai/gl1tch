// Package boards описывает доски (разделы форума), в которых лежат темы.
package boards

import (
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

type Board struct {
	ID          string
	Slug        string
	Title       string
	Description string
	Visibility  domain.Visibility // public или moderator
	SortOrder   int

	ThreadCount int
	PostCount   int

	LastPostAuthor *domain.Author
	LastPostAt     *time.Time

	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// VisibleTo показывает, может ли читатель видеть доску. Модератор видит все.
func (b Board) VisibleTo(moderator bool) bool {
	return moderator || b.Visibility == domain.VisibilityPublic
}

// LastPost последний пост для превью доски.
type LastPost struct {
	Author domain.Author
	At     time.Time
}

// Delta изменение счётчиков доски. Значения могут быть отрицательными. LastPost nil оставляет превью как есть.
type Delta struct {
	Threads  int
	Posts    int
	LastPost *LastPost
}

var (
	ErrNotFound  = apperr.NotFound("BOARD_NOT_FOUND", "board not found")
	ErrSlugTaken = apperr.AlreadyExists("BOARD_SLUG_TAKEN", "board slug is already taken")
)
