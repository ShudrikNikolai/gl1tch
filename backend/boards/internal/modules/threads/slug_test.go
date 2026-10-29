package threads

import (
	"strings"
	"testing"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
)

const testID = "0191a2b3-c4d5-7e6f-8a9b-0c1d2e3f4a5b"

func TestMakeSlug(t *testing.T) {
	cases := map[string]string{
		"Привет, мир!":    "privet-mir-2e3f4a5b",
		"Hello   World":   "hello-world-2e3f4a5b",
		"Съезд":           "sezd-2e3f4a5b",
		"Щука и Ёж":       "shchuka-i-ezh-2e3f4a5b",
		"  --Go 1.26--  ": "go-1-26-2e3f4a5b",
		"!!!":             "2e3f4a5b",
		"":                "2e3f4a5b",
	}
	for title, want := range cases {
		if got := MakeSlug(title, testID); got != want {
			t.Errorf("MakeSlug(%q) = %q, want %q", title, got, want)
		}
	}
}

func TestMakeSlugTruncatesLongTitles(t *testing.T) {
	got := MakeSlug(strings.Repeat("слово ", 40), testID)

	base := strings.TrimSuffix(got, "-2e3f4a5b")
	if len(base) > maxSlugBase || strings.HasSuffix(base, "-") {
		t.Fatalf("bad base %q (len %d)", base, len(base))
	}
}

func TestVisibleTo(t *testing.T) {
	author := domain.Author{UserID: "u1", Username: "nik"}

	cases := []struct {
		name       string
		visibility domain.Visibility
		viewer     Viewer
		want       bool
	}{
		{"public: anyone", domain.VisibilityPublic, Viewer{}, true},
		{"author: the author", domain.VisibilityAuthor, Viewer{UserID: "u1"}, true},
		{"author: someone else", domain.VisibilityAuthor, Viewer{UserID: "u2"}, false},
		{"author: anonymous", domain.VisibilityAuthor, Viewer{}, false},
		{"author: moderator", domain.VisibilityAuthor, Viewer{UserID: "u3", Moderator: true}, true},
		{"moderator: the author", domain.VisibilityModerator, Viewer{UserID: "u1"}, false},
		{"moderator: moderator", domain.VisibilityModerator, Viewer{UserID: "u3", Moderator: true}, true},
	}
	for _, c := range cases {
		th := Thread{Visibility: c.visibility, Author: author}
		if got := th.VisibleTo(c.viewer); got != c.want {
			t.Errorf("%s: VisibleTo = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestAcceptsPosts(t *testing.T) {
	for status, want := range map[Status]bool{StatusOpen: true, StatusLocked: false, StatusArchived: false} {
		if got := (Thread{Status: status}).AcceptsPosts(); got != want {
			t.Errorf("%s: AcceptsPosts = %v, want %v", status, got, want)
		}
	}
}
