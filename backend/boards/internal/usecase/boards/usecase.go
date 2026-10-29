package boards

import (
	"context"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/uuid"
	validation "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/validate"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/access"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

// Repo то, что нужно этому usecase от репозитория досок.
type Repo interface {
	Create(ctx context.Context, b boards.Board) (boards.Board, error)
	GetByID(ctx context.Context, id string) (boards.Board, error)
	List(ctx context.Context, includeModerator bool) ([]boards.Board, error)
}

type UseCase struct {
	repo Repo
	now  func() time.Time
}

func New(repo Repo) *UseCase {
	return &UseCase{repo: repo, now: time.Now}
}

// CreateBoard только для модераторов и админов.
func (uc *UseCase) CreateBoard(ctx context.Context, actor models.Actor, in models.CreateBoardInput) (boards.Board, error) {
	if err := access.RequireModerator(actor); err != nil {
		return boards.Board{}, err
	}

	slug, err := validation.Slug(in.Slug)
	if err != nil {
		return boards.Board{}, err
	}
	title, err := validation.Title(in.Title)
	if err != nil {
		return boards.Board{}, err
	}
	description, err := validation.Description(in.Description)
	if err != nil {
		return boards.Board{}, err
	}

	visibility := in.Visibility
	if visibility == "" {
		visibility = domain.VisibilityPublic
	}
	if visibility != domain.VisibilityPublic && visibility != domain.VisibilityModerator {
		return boards.Board{}, apperr.InvalidArgument("INVALID_VISIBILITY", "board visibility must be public or moderator")
	}

	return uc.repo.Create(ctx, boards.Board{
		ID:          uuid.NewUUIDv7(uc.now()),
		Slug:        slug,
		Title:       title,
		Description: description,
		Visibility:  visibility,
		SortOrder:   in.SortOrder,
		CreatedBy:   actor.UserID,
	})
}

// ListBoards доски только для модераторов остальным не показываются.
func (uc *UseCase) ListBoards(ctx context.Context, actor models.Actor) ([]boards.Board, error) {
	if err := access.RequireAuthenticated(actor); err != nil {
		return nil, err
	}

	return uc.repo.List(ctx, actor.IsModerator())
}

// GetBoard скрытая от читателя доска выглядит как несуществующая.
func (uc *UseCase) GetBoard(ctx context.Context, actor models.Actor, id string) (boards.Board, error) {
	if err := access.RequireAuthenticated(actor); err != nil {
		return boards.Board{}, err
	}

	return access.VisibleBoard(ctx, uc.repo, actor, id)
}
