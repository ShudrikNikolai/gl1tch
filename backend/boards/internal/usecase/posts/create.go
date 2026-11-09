package posts

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/uuid"
	validation "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/validate"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/access"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
)

// CreatePost добавляет пост в открытую тему. Ответить можно только на живой пост верхнего уровня той же темы.
func (uc *UseCase) CreatePost(ctx context.Context, actor models.Actor, in models.CreatePostInput) (posts.Post, error) {
	body, err := validation.Body(in.Body)
	if err != nil {
		return posts.Post{}, err
	}

	author, err := access.Author(ctx, uc.users, actor)
	if err != nil {
		return posts.Post{}, err
	}
	thread, board, err := access.VisibleThread(ctx, uc.threads, uc.boards, actor, in.ThreadID)
	if err != nil {
		return posts.Post{}, err
	}
	if !thread.AcceptsPosts() {
		return posts.Post{}, access.ErrThreadClosed
	}
	if in.ParentID != "" {
		if err := uc.checkParent(ctx, thread.ID, in.ParentID); err != nil {
			return posts.Post{}, err
		}
	}

	now := uc.now()

	var created posts.Post
	err = uc.tx.InTx(ctx, func(ctx context.Context) error {
		number, err := uc.threads.NextPostNumber(ctx, thread.ID, author, now)
		if err != nil {
			return err
		}

		created, err = uc.posts.Create(ctx, posts.Post{
			ID:        uuid.NewUUIDv7(now),
			ThreadID:  thread.ID,
			Number:    number,
			ParentID:  in.ParentID,
			Author:    author,
			Body:      body,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return err
		}

		// превью доски обновляют только публичные темы (см. threads.CreateThread)
		delta := boards.Delta{Posts: 1}
		if thread.Visibility == domain.VisibilityPublic {
			delta.LastPost = &boards.LastPost{Author: author, At: now}
		}

		return uc.boards.Apply(ctx, board.ID, delta)
	})
	if err != nil {
		return posts.Post{}, err
	}

	// Уведомления best-effort: пост уже создан, ошибка публикации не отменяет его.
	if uc.publisher != nil {
		if err := uc.publisher.PostCreated(ctx, thread, created); err != nil {
			uc.log.WarnContext(ctx, "publish post created event failed",
				slog.String("post_id", created.ID), slog.String("thread_id", thread.ID), slog.Any("error", err))
		}
	}

	return created, nil
}

func (uc *UseCase) checkParent(ctx context.Context, threadID, parentID string) error {
	parent, err := uc.posts.GetByID(ctx, parentID)
	if errors.Is(err, posts.ErrNotFound) {
		return ErrInvalidParent
	}
	if err != nil {
		return err
	}
	// один уровень вложенности: родитель сам не может быть ответом
	if parent.ThreadID != threadID || parent.ParentID != "" || parent.Deleted() {
		return ErrInvalidParent
	}

	return nil
}
