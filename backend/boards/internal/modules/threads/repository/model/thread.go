package model

import "time"

// Thread строка таблицы threads.
type Thread struct {
	ID         string
	BoardID    string
	Title      string
	Slug       string
	AuthorID   string
	AuthorName string
	Status     string
	Category   string
	Visibility string
	Pinned     bool
	PostCount  int

	LastPostAuthorID   string
	LastPostAuthorName string
	LastPostAt         time.Time

	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
