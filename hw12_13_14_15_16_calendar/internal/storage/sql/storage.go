package sqlstorage

import (
	"context"
	"database/sql"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storage"
	_ "github.com/jackc/pgx/v4/stdlib" // Регистрация драйвера PostgreSQL "pgx" для database/sql.
)

type Storage struct {
	db     *sql.DB
	logger Logger
}

type Logger interface {
	Error(msg string)
}

func New(dsn string, logger Logger) (*Storage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error("ошибка подключения к БД: " + err.Error())
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logger.Error("пинг к БД: " + err.Error())
		return nil, err
	}

	return &Storage{db: db, logger: logger}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO events (id, title, start_at, duration, user_id, notify_before, notified)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		event.ID, event.Title, event.StartAt, event.Duration, event.UserID, event.NotifyBefore, event.Notified)
	if err != nil {
		s.logger.Error("создание события: " + err.Error())
		return err
	}

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, id int64, event storage.Event) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE events SET title = $1, start_at = $2, duration = $3, user_id = $4, notify_before = $5, notified = $6
				 WHERE id = $7`,
		event.Title, event.StartAt, event.Duration, event.UserID, event.NotifyBefore, false, id)
	if err != nil {
		s.logger.Error("обновление события: " + err.Error())
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE id = $1`, id)
	if err != nil {
		s.logger.Error("удаление события: " + err.Error())
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *Storage) ListEvents(ctx context.Context) ([]storage.Event, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, start_at, duration, user_id, notify_before, notified FROM events`)
	if err != nil {
		s.logger.Error("получение списка событий: " + err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanEvents(rows)
}

func (s *Storage) MarkNotified(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE events SET notified = true WHERE id = $1`, id)
	if err != nil {
		s.logger.Error("отметка об уведомлении: " + err.Error())
		return err
	}

	return nil
}

func (s *Storage) DeleteOldEvents(ctx context.Context, olderThan time.Time) (countRowsDeleted int64, err error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE start_at < $1`, olderThan)
	if err != nil {
		s.logger.Error("удаление старых событий: " + err.Error())
		return 0, err
	}

	n, _ := res.RowsAffected()
	return n, nil
}

func (s *Storage) SaveNotification(ctx context.Context, notification storage.Notification) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO notifications (event_id, title, date, user_id) VALUES ($1, $2, $3, $4)`,
		notification.EventID, notification.Title, notification.Date, notification.UserID)
	if err != nil {
		s.logger.Error("сохранение уведомления: " + err.Error())
		return err
	}

	return nil
}

func (s *Storage) EventsToNotify(ctx context.Context, now time.Time) ([]storage.Event, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, start_at, duration, user_id, notify_before, notified
		 FROM events
		 WHERE notified = FALSE
		   AND notify_before > 0
		   AND start_at - (notify_before / 1000000000) * INTERVAL '1 second' <= $1`, now)
	if err != nil {
		s.logger.Error("выбор событий для уведомления: " + err.Error())
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanEvents(rows)
}

func scanEvents(rows *sql.Rows) ([]storage.Event, error) {
	events := make([]storage.Event, 0)
	for rows.Next() {
		var event storage.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.StartAt, &event.Duration,
			&event.UserID, &event.NotifyBefore, &event.Notified); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}
