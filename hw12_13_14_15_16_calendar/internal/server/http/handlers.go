package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storage"
)

type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, id int64, event storage.Event) error
	DeleteEvent(ctx context.Context, id int64) error
	ListEvents(ctx context.Context) ([]storage.Event, error)
}

type Handlers struct {
	app    Application
	nextID atomic.Int64
}

func NewHandlers(app Application) *Handlers {
	return &Handlers{app: app}
}

func (h *Handlers) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var body CreateEventJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "неверный JSON: "+err.Error())
		return
	}

	event := storage.Event{
		ID:       h.nextID.Add(1),
		Title:    body.Title,
		StartAt:  body.StartAt,
		Duration: time.Duration(body.Duration),
		UserID:   body.UserId,
	}

	if err := h.app.CreateEvent(r.Context(), event); err != nil {
		h.writeStorageError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toEvent(event))
}

func (h *Handlers) UpdateEvent(w http.ResponseWriter, r *http.Request, id int64) {
	var body UpdateEventJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "неверный JSON: "+err.Error())
		return
	}

	event := storage.Event{
		ID:       id,
		Title:    body.Title,
		StartAt:  body.StartAt,
		Duration: time.Duration(body.Duration),
		UserID:   body.UserId,
	}

	if err := h.app.UpdateEvent(r.Context(), id, event); err != nil {
		h.writeStorageError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toEvent(event))
}

func (h *Handlers) DeleteEvent(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.app.DeleteEvent(r.Context(), id); err != nil {
		h.writeStorageError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) ListDayEvents(w http.ResponseWriter, r *http.Request, params ListDayEventsParams) {
	h.listEvents(w, r, func(ev storage.Event) bool {
		y1, m1, d1 := ev.StartAt.UTC().Date()
		y2, m2, d2 := params.Date.UTC().Date()
		return y1 == y2 && m1 == m2 && d1 == d2
	})
}

func (h *Handlers) ListWeekEvents(w http.ResponseWriter, r *http.Request, params ListWeekEventsParams) {
	h.listWeekMonth(w, r, params.Date, 7, 0)
}

func (h *Handlers) ListMonthEvents(w http.ResponseWriter, r *http.Request, params ListMonthEventsParams) {
	h.listWeekMonth(w, r, params.Date, 0, 1)
}

func (h *Handlers) listWeekMonth(w http.ResponseWriter, r *http.Request, date time.Time, days, months int) {
	start := date.UTC()
	end := start.AddDate(0, months, days)

	h.listEvents(w, r, func(ev storage.Event) bool {
		t := ev.StartAt.UTC()
		return !t.Before(start) && t.Before(end)
	})
}

func (h *Handlers) listEvents(w http.ResponseWriter, r *http.Request, keep func(storage.Event) bool) {
	events, err := h.app.ListEvents(r.Context())
	if err != nil {
		h.writeStorageError(w, err)
		return
	}

	result := make([]Event, 0, len(events))
	for _, ev := range events {
		if keep(ev) {
			result = append(result, toEvent(ev))
		}
	}

	writeJSON(w, http.StatusOK, result)
}

func toEvent(ev storage.Event) Event {
	return Event{
		Id:       ev.ID,
		Title:    ev.Title,
		StartAt:  ev.StartAt,
		Duration: int64(ev.Duration),
		UserId:   ev.UserID,
	}
}

func (h *Handlers) writeStorageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrDateBusy):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, storage.ErrEventNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "внутренняя error: "+err.Error())
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, Error{Error: msg})
}
