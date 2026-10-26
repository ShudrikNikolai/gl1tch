package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens"
)

type usedMark struct {
	userID string
	at     time.Time
}

type memRepo struct {
	sessions map[string]refresh_tokens.Session
	used     map[string]usedMark
}

func newMemRepo() *memRepo {
	return &memRepo{sessions: map[string]refresh_tokens.Session{}, used: map[string]usedMark{}}
}

func (m *memRepo) Save(_ context.Context, hash string, s refresh_tokens.Session, _ time.Duration) error {
	m.sessions[hash] = s
	return nil
}

func (m *memRepo) Get(_ context.Context, hash string) (refresh_tokens.Session, error) {
	s, ok := m.sessions[hash]
	if !ok {
		return refresh_tokens.Session{}, refresh_tokens.ErrSessionNotFound
	}
	return s, nil
}

func (m *memRepo) Delete(_ context.Context, hash, _ string) error {
	delete(m.sessions, hash)
	return nil
}

func (m *memRepo) DeleteAllByUser(_ context.Context, userID string) error {
	for h, s := range m.sessions {
		if s.UserID == userID {
			delete(m.sessions, h)
		}
	}
	return nil
}

func (m *memRepo) MarkUsed(_ context.Context, hash, userID string, at time.Time, _ time.Duration) (bool, error) {
	if _, ok := m.used[hash]; ok {
		return false, nil
	}
	m.used[hash] = usedMark{userID: userID, at: at}
	return true, nil
}

func (m *memRepo) Used(_ context.Context, hash string) (string, time.Time, bool, error) {
	u, ok := m.used[hash]
	return u.userID, u.at, ok, nil
}

type fakeRevoker struct{ users []string }

func (f *fakeRevoker) RevokeUser(_ context.Context, userID string, _ time.Time) error {
	f.users = append(f.users, userID)
	return nil
}

const grace = 5 * time.Second

func newService() (*Service, *memRepo, *fakeRevoker, *time.Time) {
	repo := newMemRepo()
	rev := &fakeRevoker{}
	clock := time.Unix(1_800_000_000, 0)

	s := New(repo, rev, 24*time.Hour, grace, slog.New(slog.NewTextHandler(io.Discard, nil)))
	s.now = func() time.Time { return clock }

	return s, repo, rev, &clock
}

func TestRotateHappyPath(t *testing.T) {
	s, repo, _, _ := newService()
	ctx := context.Background()

	first, err := s.Issue(ctx, "u1")
	if err != nil {
		t.Fatal(err)
	}

	userID, next, err := s.Rotate(ctx, first.Value)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if userID != "u1" {
		t.Fatalf("user = %q, want u1", userID)
	}
	if next.Value == "" || next.Value == first.Value {
		t.Fatal("next token must be new")
	}
	if _, ok := repo.sessions[hashToken(first.Value)]; ok {
		t.Fatal("old session must be deleted")
	}
	if _, ok := repo.sessions[hashToken(next.Value)]; !ok {
		t.Fatal("new session must exist")
	}
}

func TestReplayWithinGraceIsRejectedWithoutRevoking(t *testing.T) {
	s, repo, rev, _ := newService()
	ctx := context.Background()

	first, _ := s.Issue(ctx, "u1")
	_, next, err := s.Rotate(ctx, first.Value)
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := s.Rotate(ctx, first.Value); !errors.Is(err, refresh_tokens.ErrInvalidToken) {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}
	if len(rev.users) != 0 {
		t.Fatal("a race inside the grace window must not revoke anything")
	}
	if _, ok := repo.sessions[hashToken(next.Value)]; !ok {
		t.Fatal("the new token must stay valid")
	}
}

func TestReplayAfterGraceRevokesEverything(t *testing.T) {
	s, repo, rev, clock := newService()
	ctx := context.Background()

	first, _ := s.Issue(ctx, "u1")
	_, next, err := s.Rotate(ctx, first.Value)
	if err != nil {
		t.Fatal(err)
	}

	*clock = clock.Add(grace + time.Second)

	if _, _, err := s.Rotate(ctx, first.Value); !errors.Is(err, refresh_tokens.ErrTokenReused) {
		t.Fatalf("err = %v, want ErrTokenReused", err)
	}
	if len(rev.users) != 1 || rev.users[0] != "u1" {
		t.Fatalf("revoked = %v, want [u1]", rev.users)
	}
	if _, ok := repo.sessions[hashToken(next.Value)]; ok {
		t.Fatal("all sessions of the user must be gone")
	}
}

func TestRotateUnknownAndEmpty(t *testing.T) {
	s, _, rev, _ := newService()
	ctx := context.Background()

	for _, raw := range []string{"", "garbage"} {
		if _, _, err := s.Rotate(ctx, raw); !errors.Is(err, refresh_tokens.ErrInvalidToken) {
			t.Fatalf("Rotate(%q) err = %v, want ErrInvalidToken", raw, err)
		}
	}
	if len(rev.users) != 0 {
		t.Fatal("unknown tokens must not revoke anything")
	}
}

func TestRevokeIsIdempotent(t *testing.T) {
	s, _, _, _ := newService()
	ctx := context.Background()

	tok, _ := s.Issue(ctx, "u1")
	if err := s.Revoke(ctx, tok.Value); err != nil {
		t.Fatal(err)
	}
	if err := s.Revoke(ctx, tok.Value); err != nil {
		t.Fatalf("second revoke: %v", err)
	}
	if err := s.Revoke(ctx, ""); err != nil {
		t.Fatalf("empty revoke: %v", err)
	}
	if _, _, err := s.Rotate(ctx, tok.Value); !errors.Is(err, refresh_tokens.ErrInvalidToken) {
		t.Fatalf("rotate after logout: err = %v, want ErrInvalidToken", err)
	}
}
