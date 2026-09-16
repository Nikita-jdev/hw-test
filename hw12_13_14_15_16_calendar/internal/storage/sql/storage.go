package sqlstorage

import (
	"context"
	"database/sql"

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
		`INSERT INTO events (id, title, start_at, duration, user_id) VALUES ($1, $2, $3, $4, $5)`,
		event.ID, event.Title, event.StartAt, event.Duration, event.UserID)
	if err != nil {
		s.logger.Error("создание события: " + err.Error())
		return err
	}

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, id int64, event storage.Event) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE events SET title = $1, start_at = $2, duration = $3, user_id = $4 WHERE id = $5`,
		event.Title, event.StartAt, event.Duration, event.UserID, id)
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
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, start_at, duration, user_id FROM events`)
	if err != nil {
		s.logger.Error("получение списка событий: " + err.Error())
		return nil, err
	}

	defer rows.Close()

	events := make([]storage.Event, 0)
	for rows.Next() {
		var event storage.Event

		if err := rows.Scan(&event.ID, &event.Title, &event.StartAt, &event.Duration, &event.UserID); err != nil {
			s.logger.Error("получение списка событий поэлементно: " + err.Error())
			return nil, err
		}

		events = append(events, event)
	}

	return events, rows.Err()
}
