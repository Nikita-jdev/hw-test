package internalhttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestMetricsEndpoint проверяет, что /metrics отдаёт метрики в формате Prometheus
// и что бизнес-сценарии попадают в метрики.
func TestMetricsEndpoint(t *testing.T) {
	router := metricsMiddleware(NewRouter(newFakeApp()))

	// Выполняем ключевой бизнес-сценарий — создание события.
	rec := doJSON(t, router, http.MethodPost, "/events", EventInput{
		Title:    "Метрика",
		StartAt:  time.Date(2024, 5, 10, 12, 0, 0, 0, time.UTC),
		Duration: int64(time.Hour),
		UserId:   1,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	// Запрашиваем эндпоинт метрик.
	metricsRec := httptest.NewRecorder()
	router.ServeHTTP(metricsRec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if metricsRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /metrics, got %d", metricsRec.Code)
	}

	body := metricsRec.Body.String()
	for _, want := range []string{
		"calendar_http_requests_total",
		"calendar_http_request_duration_seconds",
		"calendar_events_created_total",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics output does not contain %q", want)
		}
	}
}
