package internalhttp

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/metrics"
	"github.com/go-chi/chi/v5"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		ip := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ip = xff
		}

		userAgent := r.UserAgent()

		info := fmt.Sprintf("%s [%s] %s %s %s %d %s %q",
			ip,
			start.Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.RequestURI,
			r.Proto,
			rec.status,
			time.Since(start),
			userAgent,
		)

		logger.Info(info)
	})
}

// metricsMiddleware собирает метрики по всем HTTP-запросам к API:
// количество запросов/ошибок и гистограмму времени обработки.
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Сам эндпоинт /metrics не инструментируем, чтобы не засорять статистику.
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		// Для группировки используем шаблон маршрута (например, /events/{id}),
		// а не конкретный URL, чтобы не порождать бесконечную кардинальность.
		path := r.URL.Path
		if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil && routeCtx.RoutePattern() != "" {
			path = routeCtx.RoutePattern()
		}

		status := strconv.Itoa(rec.status)

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
		if rec.status >= http.StatusBadRequest {
			metrics.HTTPErrorsTotal.WithLabelValues(r.Method, path, status).Inc()
		}
	})
}
