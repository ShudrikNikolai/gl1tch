package threadpg

import (
	"strings"
	"testing"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
)

func placeholders(sql string) int {
	highest := 0
	for i := 0; i < len(sql); i++ {
		if sql[i] != '$' {
			continue
		}
		n := 0
		for j := i + 1; j < len(sql) && sql[j] >= '0' && sql[j] <= '9'; j++ {
			n = n*10 + int(sql[j]-'0')
		}
		if n > highest {
			highest = n
		}
	}

	return highest
}

func TestBuildListFirstPage(t *testing.T) {
	q, args := buildList(threads.ListFilter{BoardID: "b1", Viewer: threads.Viewer{UserID: "u1"}}, 21, nil)

	if !strings.Contains(q, "ORDER BY pinned DESC, last_post_at DESC, thread_id DESC LIMIT $3") {
		t.Fatalf("unexpected order/limit: %s", q)
	}
	if strings.Contains(q, "(pinned,") {
		t.Fatalf("first page must not have a cursor condition: %s", q)
	}
	if placeholders(q) != len(args) {
		t.Fatalf("placeholders %d != args %d", placeholders(q), len(args))
	}
	if args[len(args)-1] != 21 {
		t.Fatalf("limit arg = %v", args[len(args)-1])
	}
}

func TestBuildListCursorAndFilters(t *testing.T) {
	at := time.Date(2026, 10, 25, 12, 0, 0, 123456000, time.UTC)
	cur := &cursorData{Order: threads.OrderLastPost, Pinned: false, LastPostMicro: at.UnixMicro(), ID: "t9"}

	q, args := buildList(threads.ListFilter{
		BoardID:  "b1",
		Category: threads.CategoryFAQ,
		Status:   threads.StatusOpen,
		AuthorID: "a1",
		Search:   "50%_off\\",
		Viewer:   threads.Viewer{Moderator: true},
	}, 11, cur)

	for _, want := range []string{
		"category = $2", "status = $3", "author_id = $4::uuid", `title ILIKE $5 ESCAPE '\'`,
		"(pinned, last_post_at, thread_id) < ($6, $7::timestamptz, $8::uuid)", "LIMIT $9",
	} {
		if !strings.Contains(q, want) {
			t.Errorf("missing %q in:\n%s", want, q)
		}
	}
	if strings.Contains(q, "visibility =") { // само слово есть в списке столбцов, ищем условие
		t.Errorf("moderators must not be filtered by visibility:\n%s", q)
	}
	if placeholders(q) != len(args) {
		t.Fatalf("placeholders %d != args %d", placeholders(q), len(args))
	}
	if got := args[4]; got != `%50\%\_off\\%` {
		t.Errorf("search arg = %q, LIKE wildcards must be escaped", got)
	}
	if got := args[6].(time.Time); !got.Equal(at) {
		t.Errorf("cursor time = %v, want %v (microsecond precision)", got, at)
	}
}

func TestBuildListCreatedOrder(t *testing.T) {
	cur := &cursorData{Order: threads.OrderCreated, Pinned: true, ID: "t1"}
	q, _ := buildList(threads.ListFilter{BoardID: "b1", Order: threads.OrderCreated, Viewer: threads.Viewer{Moderator: true}}, 6, cur)

	if !strings.Contains(q, "(pinned, thread_id) < ($2, $3::uuid)") || !strings.Contains(q, "ORDER BY pinned DESC, thread_id DESC") {
		t.Fatalf("unexpected query: %s", q)
	}
}

func TestBuildListAnonymousViewerSeesNoPrivateThreads(t *testing.T) {
	_, args := buildList(threads.ListFilter{BoardID: "b1"}, 6, nil)

	if args[1] != anonymousID {
		t.Fatalf("anonymous viewer must map to the zero uuid, got %v", args[1])
	}
}
