package boards

import (
	"context"
	"errors"
	"testing"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/access"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/fake"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

var (
	user  = models.Actor{UserID: "u1", Roles: []string{"user"}}
	moder = models.Actor{UserID: "m1", Roles: []string{"user", "moder"}}
	admin = models.Actor{UserID: "a1", Roles: []string{"admin"}}
)

func newUC() (*UseCase, *fake.Boards) {
	repo := fake.NewBoards()

	return New(repo), repo
}

func TestCreateBoardPermissions(t *testing.T) {
	uc, _ := newUC()
	in := models.CreateBoardInput{Slug: "general", Title: "General"}

	if _, err := uc.CreateBoard(context.Background(), user, in); !errors.Is(err, access.ErrForbidden) {
		t.Fatalf("user: err = %v, want ErrForbidden", err)
	}
	if _, err := uc.CreateBoard(context.Background(), models.Actor{}, in); !errors.Is(err, access.ErrUnauthenticated) {
		t.Fatalf("anonymous: err = %v, want ErrUnauthenticated", err)
	}
	for _, a := range []models.Actor{moder, admin} {
		b, err := uc.CreateBoard(context.Background(), a, models.CreateBoardInput{Slug: "b-" + a.UserID, Title: "T"})
		if err != nil {
			t.Fatalf("%s: %v", a.UserID, err)
		}
		if b.Visibility != domain.VisibilityPublic || b.CreatedBy != a.UserID || b.ID == "" {
			t.Fatalf("unexpected board: %+v", b)
		}
	}
}

func TestCreateBoardValidation(t *testing.T) {
	uc, _ := newUC()

	cases := map[string]models.CreateBoardInput{
		"INVALID_SLUG":       {Slug: "Bad Slug", Title: "T"},
		"INVALID_TITLE":      {Slug: "ok", Title: "  "},
		"INVALID_VISIBILITY": {Slug: "ok", Title: "T", Visibility: domain.VisibilityAuthor},
	}
	for code, in := range cases {
		_, err := uc.CreateBoard(context.Background(), moder, in)
		if got := apperr.From(err).Code; got != code {
			t.Errorf("code = %q (err %v), want %q", got, err, code)
		}
	}
}

func TestCreateBoardSlugTaken(t *testing.T) {
	uc, _ := newUC()
	_, _ = uc.CreateBoard(context.Background(), moder, models.CreateBoardInput{Slug: "general", Title: "A"})

	if _, err := uc.CreateBoard(context.Background(), moder, models.CreateBoardInput{Slug: "general", Title: "B"}); !errors.Is(err, boards.ErrSlugTaken) {
		t.Fatalf("err = %v, want ErrSlugTaken", err)
	}
}

func TestModeratorBoardsAreHiddenFromUsers(t *testing.T) {
	uc, _ := newUC()
	pub, _ := uc.CreateBoard(context.Background(), moder, models.CreateBoardInput{Slug: "general", Title: "G", SortOrder: 2})
	staff, _ := uc.CreateBoard(context.Background(), moder, models.CreateBoardInput{Slug: "staff", Title: "S", Visibility: domain.VisibilityModerator, SortOrder: 1})

	forUser, _ := uc.ListBoards(context.Background(), user)
	if len(forUser) != 1 || forUser[0].ID != pub.ID {
		t.Fatalf("user sees %+v", forUser)
	}
	forModer, _ := uc.ListBoards(context.Background(), moder)
	if len(forModer) != 2 || forModer[0].ID != staff.ID {
		t.Fatalf("moderator sees %+v", forModer)
	}

	if _, err := uc.GetBoard(context.Background(), user, staff.ID); !errors.Is(err, boards.ErrNotFound) {
		t.Fatalf("hidden board must look missing, err = %v", err)
	}
	if _, err := uc.GetBoard(context.Background(), moder, staff.ID); err != nil {
		t.Fatalf("moderator: %v", err)
	}
	if _, err := uc.GetBoard(context.Background(), models.Actor{}, pub.ID); !errors.Is(err, access.ErrUnauthenticated) {
		t.Fatalf("anonymous: err = %v", err)
	}
}
