// Package domain общие для модулей понятия: автор, видимость и постраничная выдача.
package domain

// Author краткая информация об авторе. Username снимок на момент записи.
type Author struct {
	UserID   string
	Username string
}

type Visibility string

const (
	VisibilityPublic    Visibility = "public"
	VisibilityAuthor    Visibility = "author"
	VisibilityModerator Visibility = "moderator"
)

func (v Visibility) Valid() bool {
	switch v {
	case VisibilityPublic, VisibilityAuthor, VisibilityModerator:
		return true
	default:
		return false
	}
}

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// PageRequest запрос страницы. Token пустой для первой страницы.
type PageRequest struct {
	Size  int
	Token string
}

// EffectiveSize размер страницы с учётом значения по умолчанию и верхней границы.
func (p PageRequest) EffectiveSize() int {
	switch {
	case p.Size <= 0:
		return DefaultPageSize
	case p.Size > MaxPageSize:
		return MaxPageSize
	default:
		return p.Size
	}
}

// PageInfo описание следующей страницы. NextToken пустой, если страниц больше нет.
type PageInfo struct {
	NextToken string
	HasMore   bool
}
