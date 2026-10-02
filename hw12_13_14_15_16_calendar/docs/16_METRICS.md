# Мониторинг Календаря (ДЗ №16)

Сервис инструментирован метриками [Prometheus](https://prometheus.io/).
Все метрики собраны в пакете [`internal/metrics`](../internal/metrics/metrics.go)
и отдаются HTTP-эндпоинтом **`GET /metrics`** в текстовом формате Prometheus.
Разворачивать сам Prometheus для сдачи не требуется — достаточно эндпоинта и
самих метрик в коде.

## Где посмотреть метрики

```bash
# сервис calendar
curl http://localhost:8080/metrics
```

Эндпоинт отдаёт как собственные метрики сервиса (`calendar_*`), так и
стандартные коллекторы Go-рантайма (`go_*`, `process_*`).

## Список метрик

### HTTP / API

| Метрика | Тип | Метки | Назначение |
|---|---|---|---|
| `calendar_http_requests_total` | counter | `method`, `path`, `status` | Общее количество запросов к API, включая ошибки |
| `calendar_http_errors_total` | counter | `method`, `path`, `status` | Запросы, завершившиеся ошибкой (статус ≥ 400) |
| `calendar_http_request_duration_seconds` | histogram | `method`, `path` | Время обработки запросов в секундах |

Метка `path` содержит **шаблон маршрута** (например, `/events/{id}`), а не
конкретный URL — это защищает от неограниченного роста кардинальности метрик.

### Бизнес-сценарии

| Метрика | Тип | Назначение |
|---|---|---|
| `calendar_events_created_total` | counter | Количество успешно созданных событий |
| `calendar_events_updated_total` | counter | Количество успешно обновлённых событий |
| `calendar_events_deleted_total` | counter | Количество успешно удалённых событий |
| `calendar_notifications_sent_total` | counter | Количество успешно отправленных уведомлений |
| `calendar_notifications_failed_total` | counter | Количество ошибок отправки уведомлений |

### Фоновые задачи

| Метрика | Тип | Метки | Назначение |
|---|---|---|---|
| `calendar_background_task_up` | gauge | `task` | Статус задачи: `1` — работает, `0` — остановлена |
| `calendar_background_task_runs_total` | counter | `task` | Итерации/обработанные сообщения |
| `calendar_background_task_errors_total` | counter | `task` | Ошибки в фоновой задаче |

Значения метки `task`:

- `scheduler` — планировщик уведомлений (статус процесса);
- `scheduler_notify` — итерация рассылки уведомлений;
- `scheduler_cleanup` — итерация удаления старых событий;
- `storer` — сервис-хранитель уведомлений (статус и обработка сообщений).

## Почему эти метрики важны

- **`http_requests_total` / `http_requests_duration_seconds`** — базовый
  «золотой сигнал» сервиса: позволяют видеть нагрузку (RPS), долю ошибок и
  латентность (`p50/p95/p99` по бакетам гистограммы).
- **`http_errors_total`** — быстро показывает рост отказов API и помогает
  отличить клиентские ошибки (`4xx`) от серверных (`5xx`).
- **`events_created/updated/deleted_total`** — отражают реальную бизнес-активность
  и позволяют заметить аномалии (например, всплеск удалений или остановку
  создания событий).
- **`notifications_sent/failed_total`** — контролируют ключевой сценарий
  уведомлений: расхождение sent и failed сигнализирует о проблемах с Kafka.
- **`background_task_up`** — обнаруживает «тихую» остановку планировщика или
  хранителя (задача жива, но не работает).
- **`background_task_runs/errors_total`** — показывают, что фоновые задачи
  действительно выполняются, и как часто они падают.

## Как использовать для анализа

- **RPS и латентность API:**
  `rate(calendar_http_requests_total[5m])` и
  `histogram_quantile(0.95, rate(calendar_http_request_duration_seconds_bucket[5m]))`.
- **Доля ошибок:**
  `sum(rate(calendar_http_errors_total[5m])) / sum(rate(calendar_http_requests_total[5m]))`.
- **Узкие места:** сравнение `histogram_quantile` по разным `path` показывает,
  какой маршрут самый медленный; `status`-метка — какой из них чаще отдаёт `5xx`.
- **Здоровье фоновых задач:**
  `calendar_background_task_up == 0` — алерт об остановке задачи;
  `rate(calendar_background_task_errors_total[5m]) > 0` — алерт об ошибках.
- **Пропускная способность уведомлений:**
  `increase(calendar_notifications_sent_total[1h])` и сравнение с
  `increase(calendar_notifications_failed_total[1h])` выявляет проблемы
  интеграции с Kafka.

## Пример подключения Prometheus (опционально)

```yaml
scrape_configs:
  - job_name: calendar
    static_configs:
      - targets: ["calendar:8080"]
```
