package broker

import "context"

type Message struct {
	Key   []byte
	Value []byte
}

type Producer interface {
	Send(ctx context.Context, message Message) error
	Close() error
}

type Consumer interface {
	Receive(ctx context.Context) (Message, error)
	Commit(ctx context.Context) error
	Close() error
}

type Handler interface {
	Handle(ctx context.Context, message Message) error
}
