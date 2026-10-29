package posts

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/access"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/fake"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

var (
	nik    = models.Actor{UserID: "u-nik", Roles: []string{"user"}}
	anna   = models.Actor{UserID: "u-anna", Roles: []string{"user"}}
	banned = models.Actor{UserID: "u-banned", Roles: []string{"user"}}
	moder  = models.Actor{UserID: "u-moder", Roles: []string{"user", "moder"}}
)

type fixture struct {
	uc      *UseCase
	tx      *fake.Tx
	boards  *fake.Boards
	threads *fake.Threads
	posts   *fake.Posts
	thread  threads.Thread // открытая публичная тема с одним постом nik (#1)
	first   posts.Post
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	f := &fixture{tx: &fake.Tx{}, boards: fake.NewBoards(), threads: fake.NewThreads(), posts: fake.NewPosts()}
	f.boards.Items["b1"] = boards.Board{ID: "b1", Slug: "general", Title: "General", Visibility: domain.VisibilityPublic, ThreadCount: 1, PostCount: 1}

	f.thread = threads.Thread{
		ID: "t1", BoardID: "b1", Title: "Thread", Slug: "thread-1",
		Author: domain.Author{UserID: nik.UserID, Username: "nik"}, Status: threads.StatusOpen,
		Category: threads.CategoryGeneral, Visibility: domain.VisibilityPublic, PostCount: 1,
	}
	f.threads.Seed(f.thread, 1)
	f.first = posts.Post{ID: "p1", ThreadID: "t1", Number: 1, Author: f.thread.Author, Body: "first"}
	f.posts.Items["p1"] = f.first

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

func (f *fixture) post(t *testing.T, a models.Actor, body, parent string) posts.Post {
	t.Helper()

	p, err := f.uc.CreatePost(context.Background(), a, models.CreatePostInput{ThreadID: "t1", ParentID: parent, Body: body})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	return p
}

func (f *fixture) setThread(mut func(*threads.Thread)) {
	th := f.threads.Items["t1"]
	mut(&th)
	f.threads.Items["t1"] = th
}

func code(err error) string { return apperr.From(err).Code }

func TestCreatePostHappyPath(t *testing.T) {
	f := newFixture(t)

	p := f.post(t, anna, "  hello  ", "")
	if p.Number != 2 || p.Body != "hello" || p.Author.Username != "anna" || p.ThreadID != "t1" {
		t.Fatalf("unexpected post: %+v", p)
	}
	if f.post(t, nik, "again", "").Number != 3 {
		t.Fatal("numbers must be sequential")
	}
	if f.tx.Calls != 2 {
		t.Fatalf("each post must be written in a transaction, tx calls = %d", f.tx.Calls)
	}

	th := f.threads.Items["t1"]
	if th.PostCount != 3 || th.LastPostAuthor.Username != "nik" {
		t.Fatalf("thread bookkeeping: %+v", th)
	}
	b := f.boards.Items["b1"]
	if b.PostCount != 3 || b.LastPostAuthor == nil || b.LastPostAuthor.Username != "nik" {
		t.Fatalf("board bookkeeping: %+v", b)
	}
}

func TestPostInPrivateThreadDoesNotTouchBoardPreview(t *testing.T) {
	f := newFixture(t)
	f.setThread(func(th *threads.Thread) { th.Visibility = domain.VisibilityAuthor })

	f.post(t, nik, "secret", "")

	b := f.boards.Items["b1"]
	if b.PostCount != 2 || b.LastPostAuthor != nil {
		t.Fatalf("board: %+v", b)
	}
}

func TestCreatePostRejections(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	in := models.CreatePostInput{ThreadID: "t1", Body: "hi"}

	if _, err := f.uc.CreatePost(ctx, models.Actor{}, in); !errors.Is(err, access.ErrUnauthenticated) {
		t.Errorf("anonymous: %v", err)
	}
	if _, err := f.uc.CreatePost(ctx, banned, in); !errors.Is(err, access.ErrUserBanned) {
		t.Errorf("banned: %v", err)
	}
	if _, err := f.uc.CreatePost(ctx, nik, models.CreatePostInput{ThreadID: "t1", Body: "  "}); code(err) != "INVALID_BODY" {
		t.Errorf("empty body: %v", err)
	}
	if _, err := f.uc.CreatePost(ctx, nik, models.CreatePostInput{ThreadID: "t1", Body: strings.Repeat("я", posts.MaxBodyLength+1)}); code(err) != "INVALID_BODY" {
		t.Errorf("long body: %v", err)
	}
	if _, err := f.uc.CreatePost(ctx, nik, models.CreatePostInput{ThreadID: "nope", Body: "hi"}); !errors.Is(err, threads.ErrNotFound) {
		t.Errorf("unknown thread: %v", err)
	}
	if got := len(f.posts.Items); got != 1 {
		t.Fatalf("rejected requests must not create posts, have %d", got)
	}

	f.setThread(func(th *threads.Thread) { th.Visibility = domain.VisibilityAuthor })
	if _, err := f.uc.CreatePost(ctx, anna, in); !errors.Is(err, threads.ErrNotFound) {
		t.Errorf("hidden thread must look missing: %v", err)
	}
}

func TestClosedThreadsAcceptNoPosts(t *testing.T) {
	f := newFixture(t)

	for _, status := range []threads.Status{threads.StatusLocked, threads.StatusArchived} {
		f.setThread(func(th *threads.Thread) { th.Status = status })

		for _, a := range []models.Actor{nik, moder} {
			if _, err := f.uc.CreatePost(context.Background(), a, models.CreatePostInput{ThreadID: "t1", Body: "x"}); !errors.Is(err, access.ErrThreadClosed) {
				t.Errorf("%s by %s: %v", status, a.UserID, err)
			}
		}
	}
}

func TestParentRules(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	top := f.post(t, anna, "top", "p1")
	if top.ParentID != "p1" {
		t.Fatalf("reply to the first post: %+v", top)
	}

	f.posts.Items["other"] = posts.Post{ID: "other", ThreadID: "t2", Number: 1, Body: "x"}
	deleted := f.post(t, anna, "to be deleted", "")
	if err := f.uc.DeletePost(ctx, anna, deleted.ID); err != nil {
		t.Fatal(err)
	}

	for name, parent := range map[string]string{
		"a reply cannot be a parent": top.ID,
		"parent from another thread": "other",
		"parent is deleted":          deleted.ID,
		"parent does not exist":      "ghost",
	} {
		_, err := f.uc.CreatePost(ctx, nik, models.CreatePostInput{ThreadID: "t1", ParentID: parent, Body: "reply"})
		if !errors.Is(err, ErrInvalidParent) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestUpdatePost(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	p := f.post(t, anna, "original", "")

	got, err := f.uc.UpdatePost(ctx, anna, p.ID, " edited ")
	if err != nil || got.Body != "edited" || got.EditedAt == nil {
		t.Fatalf("author edit: %v %+v", err, got)
	}
	if _, err := f.uc.UpdatePost(ctx, nik, p.ID, "hijack"); !errors.Is(err, access.ErrForbidden) {
		t.Fatalf("stranger: %v", err)
	}
	if _, err := f.uc.UpdatePost(ctx, moder, p.ID, "moderated"); err != nil {
		t.Fatalf("moderator: %v", err)
	}
	if _, err := f.uc.UpdatePost(ctx, anna, p.ID, ""); code(err) != "INVALID_BODY" {
		t.Fatalf("empty body: %v", err)
	}

	f.setThread(func(th *threads.Thread) { th.Status = threads.StatusLocked })
	if _, err := f.uc.UpdatePost(ctx, anna, p.ID, "late"); !errors.Is(err, access.ErrThreadClosed) {
		t.Fatalf("author edits in a locked thread: %v", err)
	}
	if _, err := f.uc.UpdatePost(ctx, moder, p.ID, "mod late"); err != nil {
		t.Fatalf("moderator edits in a locked thread: %v", err)
	}

	if err := f.uc.DeletePost(ctx, moder, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.uc.UpdatePost(ctx, moder, p.ID, "zombie"); !errors.Is(err, posts.ErrNotFound) {
		t.Fatalf("edit of a deleted post: %v", err)
	}
}

func TestDeletePostRules(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	p := f.post(t, anna, "mine", "")
	before := f.boards.Items["b1"].PostCount // 2

	if err := f.uc.DeletePost(ctx, nik, p.ID); !errors.Is(err, access.ErrForbidden) {
		t.Fatalf("stranger: %v", err)
	}
	if err := f.uc.DeletePost(ctx, anna, p.ID); err != nil {
		t.Fatalf("author: %v", err)
	}
	if got := f.boards.Items["b1"].PostCount; got != before-1 {
		t.Fatalf("board post_count = %d, want %d", got, before-1)
	}
	if got := f.threads.Items["t1"].PostCount; got != 1 {
		t.Fatalf("thread post_count = %d, want 1", got)
	}

	if err := f.uc.DeletePost(ctx, anna, p.ID); err != nil {
		t.Fatalf("repeat delete must be safe: %v", err)
	}
	if got := f.boards.Items["b1"].PostCount; got != before-1 {
		t.Fatalf("repeat delete changed counters: %d", got)
	}

	if err := f.uc.DeletePost(ctx, nik, "p1"); !errors.Is(err, ErrFirstPost) {
		t.Fatalf("author deletes the first post: %v", err)
	}
	if err := f.uc.DeletePost(ctx, moder, "p1"); err != nil {
		t.Fatalf("moderator deletes the first post: %v", err)
	}
}

func TestDeletedPostsStayInTheListWithoutBody(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	p := f.post(t, anna, "secret words", "")
	f.post(t, nik, "reply below", p.ID)
	if err := f.uc.DeletePost(ctx, anna, p.ID); err != nil {
		t.Fatal(err)
	}

	page, err := f.uc.ListPosts(ctx, nik, models.ListPostsInput{ThreadID: "t1"})
	if err != nil || len(page.Posts) != 3 {
		t.Fatalf("list: %v %d posts", err, len(page.Posts))
	}
	gone := page.Posts[1]
	if !gone.Deleted() || gone.PublicBody() != "" || gone.Body != "secret words" {
		t.Fatalf("deleted post: %+v (body must stay in storage and be hidden from readers)", gone)
	}
	if page.Posts[2].ParentID != p.ID {
		t.Fatal("the reply under a deleted post must stay in place")
	}
}

func TestReadsFromHiddenThreadsLookMissing(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.setThread(func(th *threads.Thread) { th.Visibility = domain.VisibilityAuthor })

	if _, err := f.uc.ListPosts(ctx, anna, models.ListPostsInput{ThreadID: "t1"}); !errors.Is(err, threads.ErrNotFound) {
		t.Errorf("list: %v", err)
	}
	if _, err := f.uc.GetPost(ctx, anna, "p1"); !errors.Is(err, posts.ErrNotFound) {
		t.Errorf("get: %v", err)
	}
	if _, err := f.uc.GetPost(ctx, nik, "p1"); err != nil {
		t.Errorf("the author must see it: %v", err)
	}

	f.setThread(func(th *threads.Thread) { th.Visibility = domain.VisibilityPublic })
	f.boards.Items["b1"] = withVisibility(f.boards.Items["b1"], domain.VisibilityModerator)
	if _, err := f.uc.GetPost(ctx, anna, "p1"); !errors.Is(err, posts.ErrNotFound) {
		t.Errorf("post on a staff board: %v", err)
	}
}

func withVisibility(b boards.Board, v domain.Visibility) boards.Board {
	b.Visibility = v

	return b
}
