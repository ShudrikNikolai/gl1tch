package model

import "time"

// Board строка таблицы boards.
type Board struct {
	ID          string
	Slug        string
	Title       string
	Description string
	Visibility  string
	SortOrder   int

	ThreadCount int
	PostCount   int

	LastPostAuthorID   *string
	LastPostAuthorName *string
	LastPostAt         *time.Time

	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}
