package internalhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storage"
)

// fakeApp — заглушка приложения, работает в памяти.
type fakeApp struct {
	events map[int64]storage.Event
	nextID int64
}

func newFakeApp() *fakeApp {
	return &fakeApp{events: make(map[int64]storage.Event)}
}

func (f *fakeApp) CreateEvent(_ context.Context, event storage.Event) error {
	f.nextID++
	event.ID = f.nextID
	f.events[event.ID] = event
	return nil
}

func (f *fakeApp) UpdateEvent(_ context.Context, id int64, event storage.Event) error {
	if _, ok := f.events[id]; !ok {
		return storage.ErrEventNotFound
	}
	event.ID = id
	f.events[id] = event
	return nil
}

func (f *fakeApp) DeleteEvent(_ context.Context, id int64) error {
	if _, ok := f.events[id]; !ok {
		return storage.ErrEventNotFound
	}
	delete(f.events, id)
	return nil
}

func (f *fakeApp) ListEvents(_ context.Context) ([]storage.Event, error) {
	events := make([]storage.Event, 0, len(f.events))
	for _, ev := range f.events {
		events = append(events, ev)
	}
	return events, nil
}

// newTestRouter строит роутер с фейковым приложением.
func newTestRouter() http.Handler {
	app := newFakeApp()
	return Handler(NewHandlers(app))
}

func doJSON(t *testing.T, router http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestCreateEvent(t *testing.T) {
	router := newTestRouter()

	rec := doJSON(t, router, http.MethodPost, "/events", EventInput{
		Title:    "Встреча",
		StartAt:  time.Date(2024, 5, 10, 12, 0, 0, 0, time.UTC),
		Duration: int64(time.Hour),
		UserId:   1,
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	var ev Event
	if err := json.NewDecoder(rec.Body).Decode(&ev); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if ev.Title != "Встреча" || ev.Id == 0 {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestUpdateAndDeleteEvent(t *testing.T) {
	router := newTestRouter()

	rec := doJSON(t, router, http.MethodPost, "/events", EventInput{
		Title:    "Было",
		StartAt:  time.Date(2024, 5, 10, 12, 0, 0, 0, time.UTC),
		Duration: int64(time.Hour),
		UserId:   1,
	})
	var created Event
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}

	// обновление
	rec = doJSON(t, router, http.MethodPut, "/events/"+strconv.FormatInt(created.Id, 10), EventInput{
		Title:    "Стало",
		StartAt:  created.StartAt,
		Duration: created.Duration,
		UserId:   created.UserId,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d", rec.Code)
	}

	// удаление
	rec = doJSON(t, router, http.MethodDelete, "/events/"+strconv.FormatInt(created.Id, 10), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on delete, got %d", rec.Code)
	}

	// удаление несуществующего -> 404
	rec = doJSON(t, router, http.MethodDelete, "/events/9999", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestListDayEvents(t *testing.T) {
	router := newTestRouter()

	// два события: 10 мая и 11 мая
	doJSON(t, router, http.MethodPost, "/events", EventInput{
		Title:    "Первое",
		StartAt:  time.Date(2024, 5, 10, 9, 0, 0, 0, time.UTC),
		Duration: int64(time.Hour),
		UserId:   1,
	})
	doJSON(t, router, http.MethodPost, "/events", EventInput{
		Title:    "Второе",
		StartAt:  time.Date(2024, 5, 11, 9, 0, 0, 0, time.UTC),
		Duration: int64(time.Hour),
		UserId:   1,
	})

	date := "2024-05-10T00:00:00Z"
	rec := doJSON(t, router, http.MethodGet, "/events/day?date="+date, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var list []Event
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 1 || list[0].Title != "Первое" {
		t.Fatalf("expected 1 event 'Первое', got %+v", list)
	}
}
