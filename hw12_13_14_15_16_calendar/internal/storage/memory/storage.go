package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storage"
)

type Storage struct {
	mu            sync.RWMutex
	events        map[int64]storage.Event
	notifications []storage.Notification
}

func New() *Storage {
	return &Storage{events: make(map[int64]storage.Event)}
}

func (s *Storage) CreateEvent(_ context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isBusy(event, 0) {
		return storage.ErrDateBusy
	}

	s.events[event.ID] = event
	return nil
}

func (s *Storage) UpdateEvent(_ context.Context, id int64, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[id]; !ok {
		return storage.ErrEventNotFound
	}

	if s.isBusy(event, id) {
		return storage.ErrDateBusy
	}

	s.events[id] = event
	return nil
}

func (s *Storage) DeleteEvent(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[id]; !ok {
		return storage.ErrEventNotFound
	}

	delete(s.events, id)
	return nil
}

func (s *Storage) ListEvents(_ context.Context) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]storage.Event, 0, len(s.events))
	for _, event := range s.events {
		events = append(events, event)
	}

	return events, nil
}

func (s *Storage) EventsToNotify(_ context.Context, now time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]storage.Event, 0)
	for _, event := range s.events {
		if event.Notified || event.NotifyBefore <= 0 {
			continue
		}
		if !event.StartAt.Add(-event.NotifyBefore).After(now) {
			events = append(events, event)
		}
	}

	return events, nil
}

func (s *Storage) MarkNotified(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.events[id]
	if !ok {
		return storage.ErrEventNotFound
	}

	event.Notified = true
	s.events[id] = event
	return nil
}

func (s *Storage) DeleteOldEvents(_ context.Context, olderThan time.Time) (countRowsDeleted int64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var deleted int64
	for id, event := range s.events {
		if event.StartAt.Before(olderThan) {
			delete(s.events, id)
			deleted++
		}
	}
	return deleted, nil
}

func (s *Storage) SaveNotification(_ context.Context, notification storage.Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.notifications = append(s.notifications, notification)
	return nil
}

func (s *Storage) isBusy(event storage.Event, id int64) bool {
	for _, ev := range s.events {
		if id != 0 && id == ev.ID {
			continue
		}

		if event.StartAt.Before(ev.StartAt.Add(ev.Duration)) &&
			ev.StartAt.Before(event.StartAt.Add(event.Duration)) {
			return true
		}
	}
	return false
}
