package threads

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/uuid"
	validation "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/validate"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/access"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
)

// CreateThread создаёт тему и её первый пост в одной транзакции и обновляет счётчики доски.
func (uc *UseCase) CreateThread(ctx context.Context, actor models.Actor, in models.CreateThreadInput) (models.CreateThreadResult, error) {
	title, err := validation.Title(in.Title)
	if err != nil {
		return models.CreateThreadResult{}, err
	}
	body, err := validation.Body(in.Body)
	if err != nil {
		return models.CreateThreadResult{}, err
	}

	category, visibility := in.Category, in.Visibility
	if category == "" {
		category = threads.CategoryGeneral
	}
	if visibility == "" {
		visibility = domain.VisibilityPublic
	}
	if err := checkCategory(actor, category); err != nil {
		return models.CreateThreadResult{}, err
	}
	if err := checkVisibility(visibility); err != nil {
		return models.CreateThreadResult{}, err
	}

	author, err := access.Author(ctx, uc.users, actor)
	if err != nil {
		return models.CreateThreadResult{}, err
	}
	board, err := access.VisibleBoard(ctx, uc.boards, actor, in.BoardID)
	if err != nil {
		return models.CreateThreadResult{}, err
	}

	now := uc.now()
	threadID := uuid.NewUUIDv7(now)

	var res models.CreateThreadResult
	err = uc.tx.InTx(ctx, func(ctx context.Context) error {
		created, err := uc.threads.Create(ctx, threads.Thread{
			ID:             threadID,
			BoardID:        board.ID,
			Title:          title,
			Slug:           threads.MakeSlug(title, threadID),
			Author:         author,
			Status:         threads.StatusOpen,
			Category:       category,
			Visibility:     visibility,
			LastPostAuthor: author,
			LastPostAt:     now,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
		if err != nil {
			return err
		}

		number, err := uc.threads.NextPostNumber(ctx, created.ID, author, now)
		if err != nil {
			return err
		}
		post, err := uc.posts.Create(ctx, posts.Post{
			ID:        uuid.NewUUIDv7(now),
			ThreadID:  created.ID,
			Number:    number,
			Author:    author,
			Body:      body,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return err
		}

		// Счётчики доски считают все темы, а превью «последнего поста» обновляют только публичные:
		// иначе доска раскрыла бы автора и время скрытой темы.
		delta := boards.Delta{Threads: 1, Posts: 1}
		if visibility == domain.VisibilityPublic {
			delta.LastPost = &boards.LastPost{Author: author, At: now}
		}
		if err := uc.boards.Apply(ctx, board.ID, delta); err != nil {
			return err
		}

		// NextPostNumber уже записал это в БД, повторно тему не читаем
		created.PostCount = 1

		res = models.CreateThreadResult{Thread: created, FirstPost: post}

		return nil
	})
	if err != nil {
		return models.CreateThreadResult{}, err
	}

	return res, nil
}
