package posts

import (
	"context"

	validation "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/validate"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/access"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
)

// UpdatePost: текст меняет автор или модератор. Автор может править только в открытой теме и пока он не заблокирован.
func (uc *UseCase) UpdatePost(ctx context.Context, actor models.Actor, id, body string) (posts.Post, error) {
	body, err := validation.Body(body)
	if err != nil {
		return posts.Post{}, err
	}
	if _, err := access.Author(ctx, uc.users, actor); err != nil {
		return posts.Post{}, err
	}

	p, err := uc.posts.GetByID(ctx, id)
	if err != nil {
		return posts.Post{}, err
	}
	if p.Deleted() {
		return posts.Post{}, posts.ErrNotFound
	}

	thread, _, err := access.VisibleThread(ctx, uc.threads, uc.boards, actor, p.ThreadID)
	if err != nil {
		return posts.Post{}, posts.ErrNotFound
	}
	if !access.CanChange(actor, p.Author.UserID) {
		return posts.Post{}, access.ErrForbidden
	}
	if !actor.IsModerator() && !thread.AcceptsPosts() {
		return posts.Post{}, access.ErrThreadClosed
	}

	return uc.posts.UpdateBody(ctx, p.ID, body, uc.now())
}

// DeletePost: автор или модератор. Первый пост темы автор отдельно удалить не может (удаляется вся тема).
// Повторное удаление безопасно и счётчики второй раз не меняет.
func (uc *UseCase) DeletePost(ctx context.Context, actor models.Actor, id string) error {
	if err := access.RequireAuthenticated(actor); err != nil {
		return err
	}

	p, err := uc.posts.GetByID(ctx, id)
	if err != nil {
		return err
	}
	thread, board, err := access.VisibleThread(ctx, uc.threads, uc.boards, actor, p.ThreadID)
	if err != nil {
		return posts.ErrNotFound
	}
	if !access.CanChange(actor, p.Author.UserID) {
		return access.ErrForbidden
	}
	if p.Number == 1 && !actor.IsModerator() {
		return ErrFirstPost
	}

	now := uc.now()

	return uc.tx.InTx(ctx, func(ctx context.Context) error {
		_, deleted, err := uc.posts.SoftDelete(ctx, p.ID, actor.UserID, now)
		if err != nil {
			return err
		}
		if !deleted {
			return nil
		}
		if err := uc.threads.ReleasePost(ctx, thread.ID, now); err != nil {
			return err
		}

		return uc.boards.Apply(ctx, board.ID, boards.Delta{Posts: -1})
	})
}
