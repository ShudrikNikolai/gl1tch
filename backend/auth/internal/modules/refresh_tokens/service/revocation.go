package service

import (
	"context"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/cache"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/jwtx"
)

// Publisher отправляет события отзыва в Redis pub/sub (их читает jwtx.Revocations на стороне потребителей).
type Publisher struct {
	ps      cache.PubSub
	channel string
}

func NewPublisher(ps cache.PubSub, channel string) *Publisher {
	return &Publisher{ps: ps, channel: channel}
}

func (p *Publisher) RevokeUser(ctx context.Context, userID string, before time.Time) error {
	return jwtx.PublishRevocation(ctx, p.ps, p.channel, userID, before)
}
