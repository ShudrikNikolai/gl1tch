package models

import (
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
)

type CreateBoardInput struct {
	Slug        string
	Title       string
	Description string
	Visibility  domain.Visibility // пусто = public
	SortOrder   int
}

type CreateThreadInput struct {
	BoardID    string
	Title      string
	Body       string            // первый пост
	Category   threads.Category  // пусто = general
	Visibility domain.Visibility // пусто = public
}

type CreateThreadResult struct {
	Thread    threads.Thread
	FirstPost posts.Post
}

type ListThreadsInput struct {
	BoardID  string
	Category threads.Category // пусто = любая
	Status   threads.Status   // пусто = любой
	Search   string
	AuthorID string
	Order    threads.Order // пусто = last_post
	Page     domain.PageRequest
}

type ThreadsPage struct {
	Threads []threads.Thread
	Page    domain.PageInfo
}

// UpdateThreadInput пустое поле означает «не менять».
type UpdateThreadInput struct {
	ThreadID   string
	Title      string
	Category   threads.Category
	Visibility domain.Visibility
}

type CreatePostInput struct {
	ThreadID string
	ParentID string // пусто = ответ на тему
	Body     string
}

type ListPostsInput struct {
	ThreadID string
	Page     domain.PageRequest
}

type PostsPage struct {
	Posts []posts.Post
	Page  domain.PageInfo
}
