package threads

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/access"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/fake"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/enum"
)

var (
	nik    = models.Actor{UserID: "u-nik", Roles: []enum.Role{enum.RoleUser}}
	anna   = models.Actor{UserID: "u-anna", Roles: []enum.Role{enum.RoleUser}}
	banned = models.Actor{UserID: "u-banned", Roles: []enum.Role{enum.RoleUser}}
	moder  = models.Actor{UserID: "u-moder", Roles: []enum.Role{enum.RoleUser, enum.RoleModerator}}
)

type fixture struct {
	uc      *UseCase
	tx      *fake.Tx
	boards  *fake.Boards
	threads *fake.Threads
	posts   *fake.Posts
	pub     boards.Board // публичная доска
	staff   boards.Board // доска только для модераторов
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	f := &fixture{tx: &fake.Tx{}, boards: fake.NewBoards(), threads: fake.NewThreads(), posts: fake.NewPosts()}
	f.pub = boards.Board{ID: "b-pub", Slug: "general", Title: "General", Visibility: domain.VisibilityPublic}
	f.staff = boards.Board{ID: "b-staff", Slug: "staff", Title: "Staff", Visibility: domain.VisibilityModerator}
	f.boards.Items[f.pub.ID], f.boards.Items[f.staff.ID] = f.pub, f.staff

	users := fake.NewUsers(
		models.User{ID: nik.UserID, Username: "nik", Status: models.UserStatusActive},
		models.User{ID: anna.UserID, Username: "anna", Status: models.UserStatusActive},
		models.User{ID: banned.UserID, Username: "banned", Status: models.UserStatusBanned},
		models.User{ID: moder.UserID, Username: "moder", Status: models.UserStatusActive},
	)
	f.uc = New(Deps{
		Tx: f.tx, Boards: f.boards, Threads: f.threads, Posts: f.posts, Users: users,
		Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	return f
}

func (f *fixture) create(t *testing.T, a models.Actor, in models.CreateThreadInput) models.CreateThreadResult {
	t.Helper()

	if in.BoardID == "" {
		in.BoardID = f.pub.ID
	}
	if in.Title == "" {
		in.Title = "Hello world"
	}
	if in.Body == "" {
		in.Body = "first post"
	}
	res, err := f.uc.CreateThread(context.Background(), a, in)
	if err != nil {
		t.Fatalf("create thread: %v", err)
	}

	return res
}

func code(err error) string { return apperr.From(err).Code }

func TestCreateThreadHappyPath(t *testing.T) {
	f := newFixture(t)

	res := f.create(t, nik, models.CreateThreadInput{Title: "Привет, мир!"})

	if res.Thread.Author.Username != "nik" || res.Thread.Status != threads.StatusOpen || res.Thread.Category != threads.CategoryGeneral {
		t.Fatalf("unexpected thread: %+v", res.Thread)
	}
	if !strings.HasPrefix(res.Thread.Slug, "privet-mir-") || res.Thread.PostCount != 1 {
		t.Fatalf("slug/post count: %+v", res.Thread)
	}
	if res.FirstPost.Number != 1 || res.FirstPost.Body != "first post" || res.FirstPost.ThreadID != res.Thread.ID {
		t.Fatalf("unexpected first post: %+v", res.FirstPost)
	}
	if f.tx.Calls != 1 {
		t.Fatalf("thread and post must be written in one transaction, tx calls = %d", f.tx.Calls)
	}

	b := f.boards.Items[f.pub.ID]
	if b.ThreadCount != 1 || b.PostCount != 1 || b.LastPostAuthor == nil || b.LastPostAuthor.Username != "nik" {
		t.Fatalf("board counters/preview not updated: %+v", b)
	}
	if stored := f.threads.Items[res.Thread.ID]; stored.PostCount != 1 {
		t.Fatalf("stored thread post_count = %d", stored.PostCount)
	}
}

func TestCreateThreadRejections(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	cases := []struct {
		name  string
		actor models.Actor
		in    models.CreateThreadInput
		check func(error) bool
	}{
		{"anonymous", models.Actor{}, models.CreateThreadInput{BoardID: f.pub.ID, Title: "t", Body: "b"}, func(e error) bool { return errors.Is(e, access.ErrUnauthenticated) }},
		{"banned", banned, models.CreateThreadInput{BoardID: f.pub.ID, Title: "t", Body: "b"}, func(e error) bool { return errors.Is(e, access.ErrUserBanned) }},
		{"unknown board", nik, models.CreateThreadInput{BoardID: "nope", Title: "t", Body: "b"}, func(e error) bool { return errors.Is(e, boards.ErrNotFound) }},
		{"moderator-only board looks missing", nik, models.CreateThreadInput{BoardID: f.staff.ID, Title: "t", Body: "b"}, func(e error) bool { return errors.Is(e, boards.ErrNotFound) }},
		{"faq is for moderators", nik, models.CreateThreadInput{BoardID: f.pub.ID, Title: "t", Body: "b", Category: threads.CategoryFAQ}, func(e error) bool { return errors.Is(e, access.ErrForbidden) }},
		{"bad category", nik, models.CreateThreadInput{BoardID: f.pub.ID, Title: "t", Body: "b", Category: "weird"}, func(e error) bool { return code(e) == "INVALID_CATEGORY" }},
		{"bad visibility", nik, models.CreateThreadInput{BoardID: f.pub.ID, Title: "t", Body: "b", Visibility: "weird"}, func(e error) bool { return code(e) == "INVALID_VISIBILITY" }},
		{"empty title", nik, models.CreateThreadInput{BoardID: f.pub.ID, Title: " ", Body: "b"}, func(e error) bool { return code(e) == "INVALID_TITLE" }},
		{"empty body", nik, models.CreateThreadInput{BoardID: f.pub.ID, Title: "t", Body: " \n "}, func(e error) bool { return code(e) == "INVALID_BODY" }},
	}
	for _, c := range cases {
		if _, err := f.uc.CreateThread(ctx, c.actor, c.in); err == nil || !c.check(err) {
			t.Errorf("%s: err = %v", c.name, err)
		}
	}
	if len(f.threads.Items) != 0 || f.boards.Items[f.pub.ID].ThreadCount != 0 {
		t.Fatalf("rejected requests must not change anything")
	}

	if _, err := f.uc.CreateThread(ctx, moder, models.CreateThreadInput{BoardID: f.staff.ID, Title: "t", Body: "b", Category: threads.CategoryFAQ}); err != nil {
		t.Fatalf("moderator may create FAQ on a staff board: %v", err)
	}
}

func TestHiddenThreadDoesNotTouchBoardPreview(t *testing.T) {
	f := newFixture(t)

	f.create(t, nik, models.CreateThreadInput{Visibility: domain.VisibilityAuthor})

	b := f.boards.Items[f.pub.ID]
	if b.LastPostAuthor != nil || b.LastPostAt != nil {
		t.Fatalf("a private thread must not appear in the public board preview: %+v", b)
	}
	if b.ThreadCount != 1 || b.PostCount != 1 {
		t.Fatalf("counters still count it: %+v", b)
	}
}

func TestGetAndListRespectVisibility(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	public := f.create(t, nik, models.CreateThreadInput{Title: "public"}).Thread
	private := f.create(t, nik, models.CreateThreadInput{Title: "private", Visibility: domain.VisibilityAuthor}).Thread
	modOnly := f.create(t, nik, models.CreateThreadInput{Title: "for moderators", Visibility: domain.VisibilityModerator}).Thread

	if _, err := f.uc.GetThread(ctx, anna, private.ID); !errors.Is(err, threads.ErrNotFound) {
		t.Fatalf("stranger must not see an author-only thread: %v", err)
	}
	if _, err := f.uc.GetThread(ctx, nik, private.ID); err != nil {
		t.Fatalf("author: %v", err)
	}
	if _, err := f.uc.GetThread(ctx, nik, modOnly.ID); !errors.Is(err, threads.ErrNotFound) {
		t.Fatalf("even the author cannot read a moderator-only thread: %v", err)
	}
	if _, err := f.uc.GetThread(ctx, moder, modOnly.ID); err != nil {
		t.Fatalf("moderator: %v", err)
	}

	list := func(a models.Actor) []string {
		page, err := f.uc.ListThreads(ctx, a, models.ListThreadsInput{BoardID: f.pub.ID})
		if err != nil {
			t.Fatal(err)
		}
		var titles []string
		for _, th := range page.Threads {
			titles = append(titles, th.Title)
		}

		return titles
	}
	if got := list(anna); len(got) != 1 || got[0] != "public" {
		t.Fatalf("anna sees %v", got)
	}
	if got := list(nik); len(got) != 2 {
		t.Fatalf("author sees %v", got)
	}
	if got := list(moder); len(got) != 3 {
		t.Fatalf("moderator sees %v", got)
	}
	_ = public

	if _, err := f.uc.ListThreads(ctx, nik, models.ListThreadsInput{BoardID: f.staff.ID}); !errors.Is(err, boards.ErrNotFound) {
		t.Fatalf("hidden board: %v", err)
	}
	if _, err := f.uc.ListThreads(ctx, nik, models.ListThreadsInput{BoardID: f.pub.ID, Order: "updated"}); code(err) != "INVALID_ORDER" {
		t.Fatalf("unknown order: %v", err)
	}
}

func TestUpdateThreadPermissions(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	th := f.create(t, nik, models.CreateThreadInput{Title: "old"}).Thread

	got, err := f.uc.UpdateThread(ctx, nik, models.UpdateThreadInput{ThreadID: th.ID, Title: "new"})
	if err != nil || got.Title != "new" {
		t.Fatalf("author rename: %v %+v", err, got)
	}
	if _, err := f.uc.UpdateThread(ctx, anna, models.UpdateThreadInput{ThreadID: th.ID, Title: "hijack"}); !errors.Is(err, access.ErrForbidden) {
		t.Fatalf("stranger: %v", err)
	}
	if _, err := f.uc.UpdateThread(ctx, nik, models.UpdateThreadInput{ThreadID: th.ID, Visibility: domain.VisibilityAuthor}); !errors.Is(err, access.ErrForbidden) {
		t.Fatalf("author must not change visibility: %v", err)
	}
	if _, err := f.uc.UpdateThread(ctx, nik, models.UpdateThreadInput{ThreadID: th.ID, Category: threads.CategoryFAQ}); !errors.Is(err, access.ErrForbidden) {
		t.Fatalf("author must not move a thread into FAQ: %v", err)
	}
	same, err := f.uc.UpdateThread(ctx, nik, models.UpdateThreadInput{ThreadID: th.ID})
	if err != nil || same.Title != "new" {
		t.Fatalf("empty update must be a no-op: %v %+v", err, same)
	}
	if got, err := f.uc.UpdateThread(ctx, moder, models.UpdateThreadInput{ThreadID: th.ID, Visibility: domain.VisibilityModerator}); err != nil || got.Visibility != domain.VisibilityModerator {
		t.Fatalf("moderator visibility: %v %+v", err, got)
	}

	if _, err := f.uc.LockThread(ctx, moder, th.ID, ""); err != nil {
		t.Fatal(err)
	}
	f.threads.Items[th.ID] = withVisibility(f.threads.Items[th.ID], domain.VisibilityPublic)
	if _, err := f.uc.UpdateThread(ctx, nik, models.UpdateThreadInput{ThreadID: th.ID, Title: "after lock"}); !errors.Is(err, access.ErrThreadClosed) {
		t.Fatalf("author edits a locked thread: %v", err)
	}
	if _, err := f.uc.UpdateThread(ctx, moder, models.UpdateThreadInput{ThreadID: th.ID, Title: "mod edit"}); err != nil {
		t.Fatalf("moderator edits a locked thread: %v", err)
	}
}

func withVisibility(t threads.Thread, v domain.Visibility) threads.Thread {
	t.Visibility = v

	return t
}

func TestModerationIsModeratorsOnlyAndIdempotent(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	th := f.create(t, nik, models.CreateThreadInput{}).Thread

	for name, fn := range map[string]func(models.Actor) error{
		"lock":   func(a models.Actor) error { _, err := f.uc.LockThread(ctx, a, th.ID, "spam"); return err },
		"unlock": func(a models.Actor) error { _, err := f.uc.UnlockThread(ctx, a, th.ID, ""); return err },
		"pin":    func(a models.Actor) error { _, err := f.uc.PinThread(ctx, a, th.ID); return err },
		"unpin":  func(a models.Actor) error { _, err := f.uc.UnpinThread(ctx, a, th.ID); return err },
	} {
		if err := fn(nik); !errors.Is(err, access.ErrForbidden) {
			t.Errorf("%s by the author: err = %v, want ErrForbidden", name, err)
		}
		if err := fn(models.Actor{}); !errors.Is(err, access.ErrUnauthenticated) {
			t.Errorf("%s anonymous: err = %v", name, err)
		}
	}

	locked, err := f.uc.LockThread(ctx, moder, th.ID, "off topic")
	if err != nil || locked.Status != threads.StatusLocked {
		t.Fatalf("lock: %v %+v", err, locked)
	}
	if again, err := f.uc.LockThread(ctx, moder, th.ID, ""); err != nil || again.Status != threads.StatusLocked {
		t.Fatalf("second lock must be a no-op: %v", err)
	}
	if open, err := f.uc.UnlockThread(ctx, moder, th.ID, ""); err != nil || open.Status != threads.StatusOpen {
		t.Fatalf("unlock: %v %+v", err, open)
	}
	if pinned, err := f.uc.PinThread(ctx, moder, th.ID); err != nil || !pinned.Pinned {
		t.Fatalf("pin: %v %+v", err, pinned)
	}
	if unpinned, err := f.uc.UnpinThread(ctx, moder, th.ID); err != nil || unpinned.Pinned {
		t.Fatalf("unpin: %v %+v", err, unpinned)
	}
	if _, err := f.uc.LockThread(ctx, moder, th.ID, strings.Repeat("x", 501)); code(err) != "INVALID_REASON" {
		t.Fatalf("long reason: %v", err)
	}
}

func TestDeleteThreadRules(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	// автор может удалить тему, пока в ней только его пост
	alone := f.create(t, nik, models.CreateThreadInput{Title: "alone"}).Thread
	if err := f.uc.DeleteThread(ctx, anna, alone.ID); !errors.Is(err, access.ErrForbidden) {
		t.Fatalf("stranger: %v", err)
	}
	if err := f.uc.DeleteThread(ctx, nik, alone.ID); err != nil {
		t.Fatalf("author of an empty thread: %v", err)
	}
	if _, err := f.uc.GetThread(ctx, nik, alone.ID); !errors.Is(err, threads.ErrNotFound) {
		t.Fatalf("deleted thread must disappear: %v", err)
	}
	b := f.boards.Items[f.pub.ID]
	if b.ThreadCount != 0 || b.PostCount != 0 {
		t.Fatalf("board counters after delete: %+v", b)
	}

	// чужие ответы: автор уже не может, модератор может
	busy := f.create(t, nik, models.CreateThreadInput{Title: "busy"}).Thread
	f.threads.Items[busy.ID] = withPostCount(f.threads.Items[busy.ID], 4)
	f.boards.Items[f.pub.ID] = withCounters(f.boards.Items[f.pub.ID], 1, 4)

	if err := f.uc.DeleteThread(ctx, nik, busy.ID); !errors.Is(err, access.ErrForbidden) {
		t.Fatalf("author of a thread with replies: %v", err)
	}
	if err := f.uc.DeleteThread(ctx, moder, busy.ID); err != nil {
		t.Fatalf("moderator: %v", err)
	}
	b = f.boards.Items[f.pub.ID]
	if b.ThreadCount != 0 || b.PostCount != 0 {
		t.Fatalf("board counters after moderator delete: %+v", b)
	}
}

func withPostCount(t threads.Thread, n int) threads.Thread {
	t.PostCount = n

	return t
}

func withCounters(b boards.Board, threadsN, postsN int) boards.Board {
	b.ThreadCount, b.PostCount = threadsN, postsN

	return b
}
