// Package access общие для usecase проверки: кто вызывает, что ему видно и что разрешено.
package access

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

var (
	ErrUnauthenticated = apperr.Unauthenticated("UNAUTHENTICATED", "authentication required")
	ErrForbidden       = apperr.PermissionDenied("FORBIDDEN", "not enough permissions")
	ErrUserBanned      = apperr.PermissionDenied("USER_BANNED", "user is banned")
	ErrUserInactive    = apperr.Unauthenticated("USER_INACTIVE", "user is not active")
	ErrThreadClosed    = apperr.PermissionDenied("THREAD_CLOSED", "thread is closed for new posts")
)

// Users профили пользователей (реализация: клиент к сервису users).
type Users interface {
	Get(ctx context.Context, id string) (models.User, error)
}

type ThreadReader interface {
	GetByID(ctx context.Context, id string) (threads.Thread, error)
}

type BoardReader interface {
	GetByID(ctx context.Context, id string) (boards.Board, error)
}

func RequireAuthenticated(a models.Actor) error {
	if !a.Authenticated() {
		return ErrUnauthenticated
	}

	return nil
}

func RequireModerator(a models.Actor) error {
	if err := RequireAuthenticated(a); err != nil {
		return err
	}
	if !a.IsModerator() {
		return ErrForbidden
	}

	return nil
}

// CanChange: менять запись может её владелец или модератор.
func CanChange(a models.Actor, ownerID string) bool {
	return a.IsModerator() || (a.Authenticated() && a.UserID == ownerID)
}

// Author проверяет по данным из users, что пользователь активен (токен мог выдаться до бана),
// и возвращает подпись для записи: имя сохраняется в посте как снимок.
func Author(ctx context.Context, users Users, a models.Actor) (domain.Author, error) {
	if err := RequireAuthenticated(a); err != nil {
		return domain.Author{}, err
	}

	u, err := users.Get(ctx, a.UserID)
	if err != nil {
		return domain.Author{}, err
	}

	switch u.Status {
	case models.UserStatusActive:
		return domain.Author{UserID: u.ID, Username: u.Username}, nil
	case models.UserStatusBanned:
		return domain.Author{}, ErrUserBanned
	default:
		return domain.Author{}, ErrUserInactive
	}
}

// VisibleBoard загружает доску. Скрытая от читателя доска выглядит как несуществующая.
func VisibleBoard(ctx context.Context, br BoardReader, a models.Actor, id string) (boards.Board, error) {
	b, err := br.GetByID(ctx, id)
	if err != nil {
		return boards.Board{}, err
	}
	if !b.VisibleTo(a.IsModerator()) {
		return boards.Board{}, boards.ErrNotFound
	}

	return b, nil
}

// VisibleThread загружает тему вместе с доской. Если тема или её доска скрыты от читателя, это threads.ErrNotFound.
func VisibleThread(ctx context.Context, tr ThreadReader, br BoardReader, a models.Actor, id string) (threads.Thread, boards.Board, error) {
	t, err := tr.GetByID(ctx, id)
	if err != nil {
		return threads.Thread{}, boards.Board{}, err
	}
	if !t.VisibleTo(a.Viewer()) {
		return threads.Thread{}, boards.Board{}, threads.ErrNotFound
	}

	b, err := br.GetByID(ctx, t.BoardID)
	if err != nil {
		return threads.Thread{}, boards.Board{}, err
	}
	if !b.VisibleTo(a.IsModerator()) {
		return threads.Thread{}, boards.Board{}, threads.ErrNotFound
	}

	return t, b, nil
}
