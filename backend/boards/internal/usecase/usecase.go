package usecase

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
)

// Транспорту известны только эти интерфейсы. Ошибки возвращаются как apperr.

type BoardUseCase interface {
	CreateBoard(ctx context.Context, actor models.Actor, in models.CreateBoardInput) (boards.Board, error)
	ListBoards(ctx context.Context, actor models.Actor) ([]boards.Board, error)
	GetBoard(ctx context.Context, actor models.Actor, id string) (boards.Board, error)
}

type ThreadUseCase interface {
	CreateThread(ctx context.Context, actor models.Actor, in models.CreateThreadInput) (models.CreateThreadResult, error)
	ListThreads(ctx context.Context, actor models.Actor, in models.ListThreadsInput) (models.ThreadsPage, error)
	GetThread(ctx context.Context, actor models.Actor, id string) (threads.Thread, error)
	UpdateThread(ctx context.Context, actor models.Actor, in models.UpdateThreadInput) (threads.Thread, error)
	LockThread(ctx context.Context, actor models.Actor, id, reason string) (threads.Thread, error)
	UnlockThread(ctx context.Context, actor models.Actor, id, reason string) (threads.Thread, error)
	PinThread(ctx context.Context, actor models.Actor, id string) (threads.Thread, error)
	UnpinThread(ctx context.Context, actor models.Actor, id string) (threads.Thread, error)
	DeleteThread(ctx context.Context, actor models.Actor, id string) error
}

type PostUseCase interface {
	CreatePost(ctx context.Context, actor models.Actor, in models.CreatePostInput) (posts.Post, error)
	ListPosts(ctx context.Context, actor models.Actor, in models.ListPostsInput) (models.PostsPage, error)
	GetPost(ctx context.Context, actor models.Actor, id string) (posts.Post, error)
	UpdatePost(ctx context.Context, actor models.Actor, id, body string) (posts.Post, error)
	DeletePost(ctx context.Context, actor models.Actor, id string) error
	HasUserPosted(ctx context.Context, threadID, userID string) (bool, error)
}
