package kafka

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	sharedkafka "github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka"
	eventsv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/events/v1"
	"google.golang.org/protobuf/proto"
)

type fakeProducer struct {
	key   []byte
	value []byte
}

func (f *fakeProducer) Send(_ context.Context, key, value []byte, _ ...sharedkafka.Header) error {
	f.key = key
	f.value = value
	return nil
}

func TestPostCreatedEvent(t *testing.T) {
	producer := &fakeProducer{}
	pub := NewPublisher(producer, slog.New(slog.NewTextHandler(io.Discard, nil)))

	now := time.Now()
	err := pub.PostCreated(context.Background(), threads.Thread{
		ID: "thread-1", BoardID: "board-1", Title: "Hello",
	}, posts.Post{
		ID: "post-1", ThreadID: "thread-1", Number: 2, CreatedAt: now,
		Author: domain.Author{UserID: "user-1", Username: "nik"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(producer.key) != "thread-1" {
		t.Fatalf("key = %q", producer.key)
	}

	var event eventsv1.PostCreated
	if err := proto.Unmarshal(producer.value, &event); err != nil {
		t.Fatal(err)
	}
	if event.GetPostId() != "post-1" || event.GetThreadTitle() != "Hello" || event.GetPostNumber() != 2 {
		t.Fatalf("unexpected event: %+v", &event)
	}
}
