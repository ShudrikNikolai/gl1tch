package threads

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	validation "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/validate"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/access"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

// ListThreads темы доски. Темы с ограниченной видимостью и скрытые доски фильтруются под читателя.
func (uc *UseCase) ListThreads(ctx context.Context, actor models.Actor, in models.ListThreadsInput) (models.ThreadsPage, error) {
	if err := access.RequireAuthenticated(actor); err != nil {
		return models.ThreadsPage{}, err
	}

	order := in.Order
	switch order {
	case "":
		order = threads.OrderLastPost
	case threads.OrderLastPost, threads.OrderCreated:
	default:
		return models.ThreadsPage{}, apperr.InvalidArgument("INVALID_ORDER", "unknown sort order")
	}
	if in.Category != "" && !in.Category.Valid() {
		return models.ThreadsPage{}, apperr.InvalidArgument("INVALID_CATEGORY", "unknown thread category")
	}
	if in.Status != "" && !in.Status.Valid() {
		return models.ThreadsPage{}, apperr.InvalidArgument("INVALID_STATUS", "unknown thread status")
	}
	search, err := validation.Search(in.Search)
	if err != nil {
		return models.ThreadsPage{}, err
	}

	board, err := access.VisibleBoard(ctx, uc.boards, actor, in.BoardID)
	if err != nil {
		return models.ThreadsPage{}, err
	}

	items, page, err := uc.threads.List(ctx, threads.ListFilter{
		BoardID:  board.ID,
		Category: in.Category,
		Status:   in.Status,
		AuthorID: in.AuthorID,
		Search:   search,
		Order:    order,
		Viewer:   actor.Viewer(),
	}, domain.PageRequest{Size: in.Page.Size, Token: in.Page.Token})
	if err != nil {
		return models.ThreadsPage{}, err
	}

	return models.ThreadsPage{Threads: items, Page: page}, nil
}

// GetThread скрытая от читателя тема выглядит как несуществующая.
func (uc *UseCase) GetThread(ctx context.Context, actor models.Actor, id string) (threads.Thread, error) {
	if err := access.RequireAuthenticated(actor); err != nil {
		return threads.Thread{}, err
	}

	t, _, err := access.VisibleThread(ctx, uc.threads, uc.boards, actor, id)

	return t, err
}
