package grpc_api

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	boarduc "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/fake"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
	postuc "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/posts"
	threaduc "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/reqctx"
	boardsv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/boards/v1"
	commonv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/common/v1"
)

// env собирает реальные usecase поверх репозиториев в памяти: тесты проходят весь путь хендлер -> usecase -> хранилище.
type env struct {
	boards  *BoardGrpcServer
	threads *ThreadGrpcServer
	posts   *PostGrpcServer
}

func newEnv() *env {
	fb, ft, fp, tx := fake.NewBoards(), fake.NewThreads(), fake.NewPosts(), &fake.Tx{}
	users := fake.NewUsers(
		models.User{ID: "u-nik", Username: "nik", Status: models.UserStatusActive},
		models.User{ID: "u-moder", Username: "moder", Status: models.UserStatusActive},
	)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &env{
		boards:  &BoardGrpcServer{uc: boarduc.New(fb)},
		threads: &ThreadGrpcServer{uc: threaduc.New(threaduc.Deps{Tx: tx, Boards: fb, Threads: ft, Posts: fp, Users: users, Log: log})},
		posts:   &PostGrpcServer{uc: postuc.New(postuc.Deps{Tx: tx, Boards: fb, Threads: ft, Posts: fp, Users: users, Log: log})},
	}
}

// asUser имитирует то, что делает интерцептор grpcx по заголовкам x-user-id и x-roles.
func asUser(id string, roles ...string) context.Context {
	return reqctx.WithUser(context.Background(), id, roles)
}

func code(err error) string { return apperr.From(err).Code }

func TestForumFlowThroughHandlers(t *testing.T) {
	e := newEnv()
	nik, moder := asUser("u-nik", "user"), asUser("u-moder", "user", "moder")

	if _, err := e.boards.CreateBoard(nik, &boardsv1.CreateBoardRequest{Slug: "general", Title: "General"}); code(err) != "FORBIDDEN" {
		t.Fatalf("a regular user must not create boards: %v", err)
	}
	created, err := e.boards.CreateBoard(moder, &boardsv1.CreateBoardRequest{Slug: "general", Title: "General", SortOrder: 3})
	if err != nil {
		t.Fatal(err)
	}
	board := created.GetBoard()
	if board.GetVisibility() != boardsv1.Visibility_VISIBILITY_PUBLIC || board.GetSortOrder() != 3 {
		t.Fatalf("board: %+v", board)
	}

	thr, err := e.threads.CreateThread(nik, &boardsv1.CreateThreadRequest{BoardId: board.GetBoardId(), Title: "Hello", Body: "first"})
	if err != nil {
		t.Fatal(err)
	}
	th := thr.GetThread()
	if th.GetStatus() != boardsv1.ThreadStatus_THREAD_STATUS_OPEN || th.GetCategory() != boardsv1.ThreadCategory_THREAD_CATEGORY_GENERAL ||
		th.GetPostCount() != 1 || thr.GetFirstPost().GetNumber() != 1 || th.GetAuthor().GetUsername() != "nik" {
		t.Fatalf("thread: %+v first post: %+v", th, thr.GetFirstPost())
	}

	reply, err := e.posts.CreatePost(moder, &boardsv1.CreatePostRequest{ThreadId: th.GetThreadId(), ParentPostId: thr.GetFirstPost().GetPostId(), Body: "welcome"})
	if err != nil {
		t.Fatal(err)
	}
	if reply.GetPost().GetNumber() != 2 || reply.GetPost().GetParentPostId() != thr.GetFirstPost().GetPostId() {
		t.Fatalf("reply: %+v", reply.GetPost())
	}

	if _, err := e.posts.DeletePost(moder, &boardsv1.DeletePostRequest{PostId: reply.GetPost().GetPostId()}); err != nil {
		t.Fatal(err)
	}
	list, err := e.posts.ListPosts(nik, &boardsv1.ListPostsRequest{ThreadId: th.GetThreadId(), Page: &commonv1.CursorPaginationRequest{PageSize: 10}})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.GetPosts()) != 2 {
		t.Fatalf("deleted posts stay in the list: %d", len(list.GetPosts()))
	}
	if gone := list.GetPosts()[1]; !gone.GetDeleted() || gone.GetBody() != "" {
		t.Fatalf("a deleted post must be masked for readers: %+v", gone)
	}
	if list.GetPage() == nil {
		t.Fatal("the response must carry page info")
	}

	got, err := e.threads.GetThread(nik, &boardsv1.GetThreadRequest{ThreadId: th.GetThreadId()})
	if err != nil || got.GetThread().GetPostCount() != 1 {
		t.Fatalf("get thread: %v %+v", err, got.GetThread())
	}
	if _, err := e.threads.LockThread(nik, &boardsv1.LockThreadRequest{ThreadId: th.GetThreadId()}); code(err) != "FORBIDDEN" {
		t.Fatalf("lock by a regular user: %v", err)
	}
	locked, err := e.threads.LockThread(moder, &boardsv1.LockThreadRequest{ThreadId: th.GetThreadId(), Reason: "done"})
	if err != nil || locked.GetThread().GetStatus() != boardsv1.ThreadStatus_THREAD_STATUS_LOCKED {
		t.Fatalf("lock: %v %+v", err, locked.GetThread())
	}
	if _, err := e.posts.CreatePost(nik, &boardsv1.CreatePostRequest{ThreadId: th.GetThreadId(), Body: "late"}); code(err) != "THREAD_CLOSED" {
		t.Fatalf("a locked thread accepts no posts: %v", err)
	}
	if _, err := e.threads.DeleteThread(moder, &boardsv1.DeleteThreadRequest{ThreadId: th.GetThreadId()}); err != nil {
		t.Fatal(err)
	}
	boardsList, err := e.boards.ListBoards(nik, &boardsv1.ListBoardsRequest{})
	if err != nil || len(boardsList.GetBoards()) != 1 || boardsList.GetBoards()[0].GetThreadCount() != 0 {
		t.Fatalf("list boards: %v %+v", err, boardsList.GetBoards())
	}
}

func TestAnonymousCallsAreRejected(t *testing.T) {
	e := newEnv()

	if _, err := e.boards.ListBoards(context.Background(), &boardsv1.ListBoardsRequest{}); code(err) != "UNAUTHENTICATED" {
		t.Fatalf("anonymous: %v", err)
	}
}

func TestUnknownEnumValuesBecomeInvalidArgument(t *testing.T) {
	e := newEnv()
	moder := asUser("u-moder", "moder")
	b, _ := e.boards.CreateBoard(moder, &boardsv1.CreateBoardRequest{Slug: "general", Title: "G"})

	_, err := e.threads.CreateThread(moder, &boardsv1.CreateThreadRequest{
		BoardId: b.GetBoard().GetBoardId(), Title: "t", Body: "b", Category: boardsv1.ThreadCategory(99)})
	if code(err) != "INVALID_CATEGORY" {
		t.Fatalf("category 99: %v", err)
	}
	_, err = e.threads.ListThreads(moder, &boardsv1.ListThreadsRequest{BoardId: b.GetBoard().GetBoardId(), OrderBy: boardsv1.OrderByThread(99)})
	if code(err) != "INVALID_ORDER" {
		t.Fatalf("order 99: %v", err)
	}
	_, err = e.threads.ListThreads(moder, &boardsv1.ListThreadsRequest{BoardId: b.GetBoard().GetBoardId(), Status: boardsv1.ThreadStatus(99)})
	if code(err) != "INVALID_STATUS" {
		t.Fatalf("status 99: %v", err)
	}
	_, err = e.threads.CreateThread(moder, &boardsv1.CreateThreadRequest{
		BoardId: b.GetBoard().GetBoardId(), Title: "t", Body: "b", Visibility: boardsv1.Visibility(99)})
	if code(err) != "INVALID_VISIBILITY" {
		t.Fatalf("thread visibility 99: %v", err)
	}
	_, err = e.boards.CreateBoard(moder, &boardsv1.CreateBoardRequest{Slug: "other", Title: "O", Visibility: boardsv1.Visibility(99)})
	if code(err) != "INVALID_VISIBILITY" {
		t.Fatalf("board visibility 99: %v", err)
	}
}

func TestEnumMappingsRoundTrip(t *testing.T) {
	for _, v := range []domain.Visibility{domain.VisibilityPublic, domain.VisibilityAuthor, domain.VisibilityModerator} {
		if got := visibilityFromProto(visibilityToProto(v)); got != v {
			t.Errorf("visibility %s -> %s", v, got)
		}
	}
	for _, v := range []threads.Category{threads.CategoryGeneral, threads.CategoryFAQ, threads.CategoryFlood} {
		if got := categoryFromProto(categoryToProto(v)); got != v {
			t.Errorf("category %s -> %s", v, got)
		}
	}
	for _, v := range []threads.Status{threads.StatusOpen, threads.StatusLocked, threads.StatusArchived} {
		if got := statusFromProto(statusToProto(v)); got != v {
			t.Errorf("status %s -> %s", v, got)
		}
	}
	if orderFromProto(boardsv1.OrderByThread_ORDER_BY_THREAD_UNSPECIFIED) != "" ||
		orderFromProto(boardsv1.OrderByThread_ORDER_BY_THREAD_CREATED_AT) != threads.OrderCreated ||
		orderFromProto(boardsv1.OrderByThread_ORDER_BY_THREAD_LAST_POST_AT) != threads.OrderLastPost {
		t.Error("order mapping")
	}
	if visibilityFromProto(boardsv1.Visibility_VISIBILITY_UNSPECIFIED) != "" || categoryFromProto(0) != "" || statusFromProto(0) != "" {
		t.Error("UNSPECIFIED must map to the empty value (\"not set\")")
	}
	if visibilityToProto("weird") != boardsv1.Visibility_VISIBILITY_UNSPECIFIED {
		t.Error("an unknown domain value must not turn into a real enum value")
	}
}

func TestConverters(t *testing.T) {
	at := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)

	empty := boardToProto(boards.Board{ID: "b1", Visibility: domain.VisibilityPublic})
	if empty.GetLastPostAuthor() != nil || empty.GetLastPostAt() != nil {
		t.Errorf("a board without posts must not invent a last post: %+v", empty)
	}
	withPost := boardToProto(boards.Board{ID: "b1", ThreadCount: 2, PostCount: 5,
		LastPostAuthor: &domain.Author{UserID: "u1", Username: "nik"}, LastPostAt: &at})
	if withPost.GetLastPostAuthor().GetUsername() != "nik" || !withPost.GetLastPostAt().AsTime().Equal(at) || withPost.GetPostCount() != 5 {
		t.Errorf("board: %+v", withPost)
	}

	if timestamp(time.Time{}) != nil {
		t.Error("the zero time must become nil, not 1970")
	}

	deleted := postToProto(posts.Post{ID: "p1", Body: "secret", DeletedAt: &at, Number: 4})
	if deleted.GetBody() != "" || !deleted.GetDeleted() || deleted.GetNumber() != 4 {
		t.Errorf("deleted post: %+v", deleted)
	}
	live := postToProto(posts.Post{ID: "p2", Body: "visible", ParentID: "p1"})
	if live.GetBody() != "visible" || live.GetDeleted() || live.GetParentPostId() != "p1" {
		t.Errorf("live post: %+v", live)
	}

	if got := pageRequest(nil); got != (domain.PageRequest{}) {
		t.Errorf("nil page request = %+v", got)
	}
	if got := pageRequest(&commonv1.CursorPaginationRequest{PageSize: 30, PageToken: "tok"}); got.Size != 30 || got.Token != "tok" {
		t.Errorf("page request = %+v", got)
	}
	if got := pageResponse(domain.PageInfo{NextToken: "n", HasMore: true}); got.GetNextPageToken() != "n" || !got.GetHasMore() {
		t.Errorf("page response = %+v", got)
	}
}
