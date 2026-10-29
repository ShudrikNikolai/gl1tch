package posts

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/access"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
)

// ListPosts посты темы по возрастанию номера. Удалённые остаются в списке: PublicBody() у них пустой.
func (uc *UseCase) ListPosts(ctx context.Context, actor models.Actor, in models.ListPostsInput) (models.PostsPage, error) {
	if err := access.RequireAuthenticated(actor); err != nil {
		return models.PostsPage{}, err
	}

	thread, _, err := access.VisibleThread(ctx, uc.threads, uc.boards, actor, in.ThreadID)
	if err != nil {
		return models.PostsPage{}, err
	}

	items, page, err := uc.posts.ListByThread(ctx, thread.ID, in.Page)
	if err != nil {
		return models.PostsPage{}, err
	}

	return models.PostsPage{Posts: items, Page: page}, nil
}

// GetPost пост из скрытой от читателя темы выглядит как несуществующий.
func (uc *UseCase) GetPost(ctx context.Context, actor models.Actor, id string) (posts.Post, error) {
	if err := access.RequireAuthenticated(actor); err != nil {
		return posts.Post{}, err
	}

	p, err := uc.posts.GetByID(ctx, id)
	if err != nil {
		return posts.Post{}, err
	}
	if _, _, err := access.VisibleThread(ctx, uc.threads, uc.boards, actor, p.ThreadID); err != nil {
		return posts.Post{}, posts.ErrNotFound
	}

	return p, nil
}
