package consumer

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/ShudrikNikolai/gl1tch/backend/notifications/internal/hub"
	sharedkafka "github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka"
	eventsv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/events/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHandlerPublishesPostCreated(t *testing.T) {
	h := hub.New()
	client := h.Subscribe("thread-1")
	defer client.Close()

	event := &eventsv1.PostCreated{
		EventId:      "event-1",
		PostId:       "post-1",
		ThreadId:     "thread-1",
		AuthorUserId: "user-1",
		CreatedAt:    timestamppb.Now(),
	}
	value, err := proto.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	handler := NewHandler(h, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := handler.Handle(context.Background(), sharedkafka.Message{Value: value}); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-client.Events():
		if got.ID != "event-1" || got.ThreadID != "thread-1" {
			t.Fatalf("unexpected event: %+v", got)
		}
	default:
		t.Fatal("event was not published")
	}
}

func TestHandlerRejectsBrokenPayload(t *testing.T) {
	handler := NewHandler(hub.New(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := handler.Handle(context.Background(), sharedkafka.Message{Value: []byte("not-proto")}); err == nil {
		t.Fatal("expected error")
	}
}
