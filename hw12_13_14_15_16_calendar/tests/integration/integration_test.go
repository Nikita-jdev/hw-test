//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storage"
	_ "github.com/jackc/pgx/v4/stdlib" // драйвер pgx для database/sql
	kafkago "github.com/segmentio/kafka-go"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func baseURL() string { return env("CALENDAR_URL", "http://localhost:8888") }
func dbDSN() string {
	return env("DB_DSN", "postgres://calendar:calendar@localhost:5432/calendar?sslmode=disable")
}
func brokers() string { return env("KAFKA_BROKERS", "localhost:9092") }
func topic() string   { return env("KAFKA_TOPIC", "calendar.notifications") }

type event struct {
	ID       int64     `json:"id"`
	Title    string    `json:"title"`
	StartAt  time.Time `json:"start_at"`
	Duration int64     `json:"duration"`
	UserID   int64     `json:"user_id"`
}

func httpClient() *http.Client { return &http.Client{Timeout: 10 * time.Second} }

// waitForAPI ждёт, пока HTTP API Календаря начнёт отвечать.
func waitForAPI(t *testing.T) {
	t.Helper()

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := httpClient().Get(baseURL() + "/events/day?date=2000-01-01T00:00:00Z")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(time.Second)
	}

	t.Fatalf("calendar API недоступен по адресу %s", baseURL())
}

func doRaw(t *testing.T, method, path string, raw []byte) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(method, baseURL()+path, bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if raw != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := httpClient().Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, _ := io.ReadAll(resp.Body)
	return resp, data
}

func doJSON(t *testing.T, method, path string, body interface{}) (*http.Response, []byte) {
	t.Helper()

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	return doRaw(t, method, path, raw)
}

func contains(items []string, want string) bool {
	for _, it := range items {
		if it == want {
			return true
		}
	}
	return false
}

// 1) Добавление события и обработка бизнес-ошибок.
func TestCreateEventAndBusinessErrors(t *testing.T) {
	waitForAPI(t)

	// успешное создание
	resp, data := doJSON(t, http.MethodPost, "/events", event{
		Title:    "integration-create",
		StartAt:  time.Date(2042, 1, 10, 12, 0, 0, 0, time.UTC),
		Duration: int64(time.Hour),
		UserID:   1,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d, body=%s", resp.StatusCode, data)
	}

	var created event
	if err := json.Unmarshal(data, &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.ID == 0 || created.Title != "integration-create" {
		t.Fatalf("unexpected created event: %+v", created)
	}

	// бизнес-ошибка: невалидный JSON -> 400
	resp, data = doRaw(t, http.MethodPost, "/events", []byte(`{"title":`))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad json: expected 400, got %d, body=%s", resp.StatusCode, data)
	}

	// бизнес-ошибка: удаление несуществующего события -> 404
	resp, data = doRaw(t, http.MethodDelete, "/events/999999", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("delete missing: expected 404, got %d, body=%s", resp.StatusCode, data)
	}

	// бизнес-ошибка: обновление несуществующего события -> 404
	resp, data = doJSON(t, http.MethodPut, "/events/999999", event{
		Title:    "no-such-event",
		StartAt:  time.Date(2042, 1, 11, 12, 0, 0, 0, time.UTC),
		Duration: int64(time.Hour),
		UserID:   1,
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("update missing: expected 404, got %d, body=%s", resp.StatusCode, data)
	}
}

// 2) Листинг событий на день / неделю / месяц.
func TestListingDayWeekMonth(t *testing.T) {
	waitForAPI(t)

	post := func(title string, at time.Time) {
		resp, data := doJSON(t, http.MethodPost, "/events", event{
			Title:    title,
			StartAt:  at,
			Duration: int64(time.Hour),
			UserID:   2,
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create %s: expected 201, got %d, body=%s", title, resp.StatusCode, data)
		}
	}

	post("list-day", time.Date(2043, 3, 10, 10, 0, 0, 0, time.UTC))
	post("list-week", time.Date(2043, 3, 12, 10, 0, 0, 0, time.UTC))
	post("list-month", time.Date(2043, 4, 20, 10, 0, 0, 0, time.UTC))

	dayTitles := listTitles(t, "/events/day?date=2043-03-10T00:00:00Z")
	if !contains(dayTitles, "list-day") || contains(dayTitles, "list-week") {
		t.Fatalf("day list unexpected: %v", dayTitles)
	}

	weekTitles := listTitles(t, "/events/week?date=2043-03-10T00:00:00Z")
	if !contains(weekTitles, "list-day") || !contains(weekTitles, "list-week") {
		t.Fatalf("week list unexpected: %v", weekTitles)
	}
	if contains(weekTitles, "list-month") {
		t.Fatalf("week list should not contain month event: %v", weekTitles)
	}

	monthTitles := listTitles(t, "/events/month?date=2043-04-01T00:00:00Z")
	if !contains(monthTitles, "list-month") {
		t.Fatalf("month list unexpected: %v", monthTitles)
	}
}

func listTitles(t *testing.T, path string) []string {
	t.Helper()

	resp, data := doRaw(t, http.MethodGet, path, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: expected 200, got %d, body=%s", path, resp.StatusCode, data)
	}

	var events []event
	if err := json.Unmarshal(data, &events); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}

	titles := make([]string, 0, len(events))
	for _, e := range events {
		titles = append(titles, e.Title)
	}
	return titles
}

// 3) Тест Хранителя: сообщение из Kafka сохраняется в таблицу notifications.
func TestStorerSavesNotificationToDB(t *testing.T) {
	waitForAPI(t)

	db, err := sql.Open("pgx", dbDSN())
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// уникальный eventID, чтобы не пересечься с другими запусками
	eventID := time.Now().UnixNano() % 1_000_000_000
	notification := storage.Notification{
		EventID: eventID,
		Title:   "integration-notification",
		Date:    time.Date(2044, 5, 1, 9, 0, 0, 0, time.UTC),
		UserID:  3,
	}

	payload, err := json.Marshal(notification)
	if err != nil {
		t.Fatalf("marshal notification: %v", err)
	}

	writer := &kafkago.Writer{
		Addr:                   kafkago.TCP(brokers()),
		Topic:                  topic(),
		Balancer:               &kafkago.LeastBytes{},
		AllowAutoTopicCreation: true,
		WriteTimeout:           10 * time.Second,
	}
	defer func() { _ = writer.Close() }()

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		// отправляем повторно: закрывает гонку старта consumer'а (StartOffset = LastOffset)
		if err := writer.WriteMessages(ctx, kafkago.Message{Value: payload}); err != nil {
			t.Logf("kafka write: %v", err)
		}

		var count int
		err := db.QueryRowContext(ctx,
			`SELECT count(*) FROM notifications WHERE event_id = $1`, eventID).Scan(&count)
		switch {
		case err != nil:
			t.Logf("db query: %v", err)
		case count > 0:
			return
		}

		time.Sleep(2 * time.Second)
	}

	t.Fatalf("уведомление для события %d не сохранено в БД", eventID)
}
