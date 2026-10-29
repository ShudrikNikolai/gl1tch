package kafka

import (
	"context"
)

type MessageHandler func(ctx context.Context, msg Message) error

// Middleware оборачивает обработчик. В списке первый элемент самый внешний
type Middleware func(next MessageHandler) MessageHandler

type Consumer interface {
	Consume(ctx context.Context, handler MessageHandler) error
}

// Header заголовок исходящего сообщения
type Header struct {
	Key   string
	Value []byte
}

type Producer interface {
	Send(ctx context.Context, key, value []byte, headers ...Header) error
}
