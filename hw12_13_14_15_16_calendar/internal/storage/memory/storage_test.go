package memorystorage

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storage"
)

func TestCreateEvent(t *testing.T) {
	s := New()
	event := storage.Event{ID: 1, Title: "event", StartAt: time.Now(), Duration: time.Hour, UserID: 1}

	if err := s.CreateEvent(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events, err := s.ListEvents(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestCreateEventDateBusy(t *testing.T) {
	s := New()

	event1 := storage.Event{ID: 1, Title: "first", StartAt: time.Now(), Duration: time.Hour, UserID: 1}
	if err := s.CreateEvent(context.Background(), event1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// пересекается с event1
	event2 := storage.Event{ID: 2, Title: "second", StartAt: time.Now().Add(time.Minute), Duration: time.Hour, UserID: 1}
	if err := s.CreateEvent(context.Background(), event2); !errors.Is(err, storage.ErrDateBusy) {
		t.Fatalf("expected ErrDateBusy, got %v", err)
	}
}

func TestUpdateEvent(t *testing.T) {
	s := New()
	event := storage.Event{ID: 1, Title: "old", StartAt: time.Now(), Duration: time.Hour, UserID: 1}
	if err := s.CreateEvent(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	event.Title = "new"
	if err := s.UpdateEvent(context.Background(), 1, event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events, _ := s.ListEvents(context.Background())
	if len(events) != 1 || events[0].Title != "new" {
		t.Fatalf("expected updated title, got %+v", events)
	}
}

func TestUpdateEventNotFound(t *testing.T) {
	s := New()
	err := s.UpdateEvent(context.Background(), 42, storage.Event{ID: 42})
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestDeleteEvent(t *testing.T) {
	s := New()
	event := storage.Event{ID: 1, Title: "event", StartAt: time.Now(), Duration: time.Hour, UserID: 1}
	if err := s.CreateEvent(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := s.DeleteEvent(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events, _ := s.ListEvents(context.Background())
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestDeleteEventNotFound(t *testing.T) {
	s := New()
	err := s.DeleteEvent(context.Background(), 42)
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := New()
	base := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			event := storage.Event{
				ID:       int64(i),
				Title:    "event",
				StartAt:  base.Add(time.Duration(i) * time.Minute),
				Duration: time.Minute,
				UserID:   1,
			}
			_ = s.CreateEvent(context.Background(), event)
		}(i)
	}
	wg.Wait()

	events, err := s.ListEvents(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 100 {
		t.Fatalf("expected 100 events, got %d", len(events))
	}
}
