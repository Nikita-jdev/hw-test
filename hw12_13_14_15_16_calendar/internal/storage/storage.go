package storage

import "context"

type Storage interface {
	CreateEvent(ctx context.Context, event Event) error
	UpdateEvent(ctx context.Context, id int64, event Event) error
	DeleteEvent(ctx context.Context, id int64) error
	ListEvents(ctx context.Context) ([]Event, error)
}
