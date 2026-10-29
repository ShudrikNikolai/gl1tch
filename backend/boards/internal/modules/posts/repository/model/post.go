package model

import "time"

// Post строка таблицы posts.
type Post struct {
	ID         string
	ThreadID   string
	Number     int
	ParentID   *string
	AuthorID   string
	AuthorName string
	Body       string
	EditedAt   *time.Time
	DeletedAt  *time.Time
	DeletedBy  *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
