// Package posts описывает посты в темах. Ответ на пост вложен на один уровень.
package posts

import (
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

// MaxBodyLength предел длины текста в символах (совпадает с CHECK в миграции).
const MaxBodyLength = 20000

type Post struct {
	ID       string
	ThreadID string
	Number   int    // порядковый номер в теме (#17), не переиспользуется
	ParentID string // пусто: пост отвечает на тему, а не на другой пост

	Author domain.Author
	Body   string // markdown

	EditedAt  *time.Time
	DeletedAt *time.Time
	DeletedBy string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p Post) Deleted() bool {
	return p.DeletedAt != nil
}

// PublicBody текст для читателей: у удалённого поста пустой, оригинал остаётся только в БД для модераторов.
func (p Post) PublicBody() string {
	if p.Deleted() {
		return ""
	}

	return p.Body
}

var ErrNotFound = apperr.NotFound("POST_NOT_FOUND", "post not found")
