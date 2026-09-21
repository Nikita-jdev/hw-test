package storage

import (
	"context"
	"time"
)

type Storage interface {
	CreateEvent(ctx context.Context, event Event) error
	UpdateEvent(ctx context.Context, id int64, event Event) error
	DeleteEvent(ctx context.Context, id int64) error
	ListEvents(ctx context.Context) ([]Event, error)

	EventsToNotify(ctx context.Context, now time.Time) ([]Event, error)
	MarkNotified(ctx context.Context, id int64) error
	DeleteOldEvents(ctx context.Context, olderThan time.Time) (int64, error)

	SaveNotification(ctx context.Context, notification Notification) error
}
