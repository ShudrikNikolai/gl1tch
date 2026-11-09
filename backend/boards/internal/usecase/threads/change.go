package threads

import (
	"context"
	"log/slog"

	validation "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/validate"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/access"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
)

// UpdateThread: заголовок и категорию меняет автор или модератор, видимость только модератор.
// Автор может править только открытую тему. Пустые поля не меняются.
func (uc *UseCase) UpdateThread(ctx context.Context, actor models.Actor, in models.UpdateThreadInput) (threads.Thread, error) {
	if err := access.RequireAuthenticated(actor); err != nil {
		return threads.Thread{}, err
	}

	t, _, err := access.VisibleThread(ctx, uc.threads, uc.boards, actor, in.ThreadID)
	if err != nil {
		return threads.Thread{}, err
	}
	if !access.CanChange(actor, t.Author.UserID) {
		return threads.Thread{}, access.ErrForbidden
	}
	if !actor.IsModerator() && !t.AcceptsPosts() {
		return threads.Thread{}, access.ErrThreadClosed
	}

	var p threads.UpdateParams
	if in.Title != "" {
		title, err := validation.Title(in.Title)
		if err != nil {
			return threads.Thread{}, err
		}
		p.Title = &title
	}
	if in.Category != "" {
		if err := checkCategory(actor, in.Category); err != nil {
			return threads.Thread{}, err
		}
		p.Category = &in.Category
	}
	if in.Visibility != "" {
		if err := checkVisibility(in.Visibility); err != nil {
			return threads.Thread{}, err
		}
		if !actor.IsModerator() {
			return threads.Thread{}, access.ErrForbidden
		}
		p.Visibility = &in.Visibility
	}

	if p == (threads.UpdateParams{}) {
		return t, nil
	}

	return uc.threads.Update(ctx, t.ID, p, uc.now())
}

// DeleteThread: модератор удаляет любую тему, автор только свою и только пока в ней нет чужих ответов.
func (uc *UseCase) DeleteThread(ctx context.Context, actor models.Actor, id string) error {
	if err := access.RequireAuthenticated(actor); err != nil {
		return err
	}

	t, board, err := access.VisibleThread(ctx, uc.threads, uc.boards, actor, id)
	if err != nil {
		return err
	}
	if !actor.IsModerator() && (actor.UserID != t.Author.UserID || t.PostCount > 1) {
		return access.ErrForbidden
	}

	now := uc.now()

	return uc.tx.InTx(ctx, func(ctx context.Context) error {
		if err := uc.threads.SoftDelete(ctx, t.ID, actor.UserID, now); err != nil {
			return err
		}

		return uc.boards.Apply(ctx, board.ID, boards.Delta{Threads: -1, Posts: -t.PostCount})
	})
}

// LockThread переводит тему в «только чтение». Повтор безопасен.
func (uc *UseCase) LockThread(ctx context.Context, actor models.Actor, id, reason string) (threads.Thread, error) {
	return uc.setStatus(ctx, actor, id, reason, threads.StatusLocked, "thread locked")
}

// UnlockThread открывает тему снова.
func (uc *UseCase) UnlockThread(ctx context.Context, actor models.Actor, id, reason string) (threads.Thread, error) {
	return uc.setStatus(ctx, actor, id, reason, threads.StatusOpen, "thread unlocked")
}

func (uc *UseCase) setStatus(ctx context.Context, actor models.Actor, id, reason string, status threads.Status, event string) (threads.Thread, error) {
	if err := access.RequireModerator(actor); err != nil {
		return threads.Thread{}, err
	}
	reason, err := validation.Reason(reason)
	if err != nil {
		return threads.Thread{}, err
	}

	t, _, err := access.VisibleThread(ctx, uc.threads, uc.boards, actor, id)
	if err != nil {
		return threads.Thread{}, err
	}
	if t.Status == status {
		return t, nil
	}

	updated, err := uc.threads.SetStatus(ctx, t.ID, status, uc.now())
	if err != nil {
		return threads.Thread{}, err
	}
	uc.log.InfoContext(ctx, event, slog.String("thread_id", t.ID), slog.String("by", actor.UserID), slog.String("reason", reason))

	return updated, nil
}

// PinThread закрепляет тему наверху списка. Повтор безопасен.
func (uc *UseCase) PinThread(ctx context.Context, actor models.Actor, id string) (threads.Thread, error) {
	return uc.setPinned(ctx, actor, id, true)
}

func (uc *UseCase) UnpinThread(ctx context.Context, actor models.Actor, id string) (threads.Thread, error) {
	return uc.setPinned(ctx, actor, id, false)
}

func (uc *UseCase) setPinned(ctx context.Context, actor models.Actor, id string, pinned bool) (threads.Thread, error) {
	if err := access.RequireModerator(actor); err != nil {
		return threads.Thread{}, err
	}

	t, _, err := access.VisibleThread(ctx, uc.threads, uc.boards, actor, id)
	if err != nil {
		return threads.Thread{}, err
	}
	if t.Pinned == pinned {
		return t, nil
	}

	return uc.threads.SetPinned(ctx, t.ID, pinned, uc.now())
}
