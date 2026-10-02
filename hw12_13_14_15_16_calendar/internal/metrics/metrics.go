// Package metrics содержит набор метрик Prometheus, которыми инструментирован
// сервис «Календарь». Все метрики автоматически регистрируются в реестре
// prometheus.DefaultRegisterer и отдаются по HTTP-эндпоинту /metrics.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal — количество запросов к API с разбивкой по методу,
	// маршруту и коду ответа (включая ошибки).
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "calendar_http_requests_total",
		Help: "Общее количество HTTP-запросов к API календаря.",
	}, []string{"method", "path", "status"})

	// HTTPErrorsTotal — количество запросов, завершившихся ошибкой (статус >= 400).
	HTTPErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "calendar_http_errors_total",
		Help: "Количество HTTP-запросов к API, завершившихся ошибкой.",
	}, []string{"method", "path", "status"})

	// HTTPRequestDuration — гистограмма времени обработки запросов в секундах.
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "calendar_http_request_duration_seconds",
		Help:    "Время обработки HTTP-запросов в секундах.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	// Бизнес-метрики по событиям календаря.
	EventsCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "calendar_events_created_total",
		Help: "Количество успешно созданных событий.",
	})

	EventsUpdatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "calendar_events_updated_total",
		Help: "Количество успешно обновлённых событий.",
	})

	EventsDeletedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "calendar_events_deleted_total",
		Help: "Количество успешно удалённых событий.",
	})

	// Метрики отправки уведомлений планировщиком.
	NotificationsSentTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "calendar_notifications_sent_total",
		Help: "Количество успешно отправленных уведомлений о событиях.",
	})

	NotificationsFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "calendar_notifications_failed_total",
		Help: "Количество ошибок при отправке уведомлений о событиях.",
	})

	// BackgroundTaskUp — статус фоновых задач: 1 — работает, 0 — остановлена.
	BackgroundTaskUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "calendar_background_task_up",
		Help: "Статус работы фоновой задачи: 1 — работает, 0 — остановлена.",
	}, []string{"task"})

	// BackgroundTaskRunsTotal — количество итераций/обработанных сообщений фоновой задачи.
	BackgroundTaskRunsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "calendar_background_task_runs_total",
		Help: "Количество итераций или обработанных сообщений фоновой задачи.",
	}, []string{"task"})

	// BackgroundTaskErrorsTotal — количество ошибок в фоновой задаче.
	BackgroundTaskErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "calendar_background_task_errors_total",
		Help: "Количество ошибок при работе фоновой задачи.",
	}, []string{"task"})
)
