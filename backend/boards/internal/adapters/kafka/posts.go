// Package kafka публикует события boards в Kafka.
package kafka

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	sharedkafka "github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/kafka"
	eventsv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/events/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Publisher struct {
	producer sharedkafka.Producer
	log      *slog.Logger
}

func NewPublisher(producer sharedkafka.Producer, log *slog.Logger) *Publisher {
	return &Publisher{producer: producer, log: log}
}

func (p *Publisher) PostCreated(ctx context.Context, thread threads.Thread, post posts.Post) error {
	if p == nil || p.producer == nil {
		return nil
	}

	event := &eventsv1.PostCreated{
		EventId:      post.ID,
		PostId:       post.ID,
		ThreadId:     thread.ID,
		BoardId:      thread.BoardID,
		ThreadTitle:  thread.Title,
		AuthorUserId: post.Author.UserID,
		AuthorName:   post.Author.Username,
		PostNumber:   int32(post.Number),
		CreatedAt:    timestamppb.New(post.CreatedAt),
	}

	value, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("kafka: marshal post created: %w", err)
	}

	return p.producer.Send(ctx, []byte(thread.ID), value)
}
