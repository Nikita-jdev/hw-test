package scheduler

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/broker"
	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storage"
)

type Scheduler struct {
	logger   Logger
	storage  Storage
	producer broker.Producer
	interval time.Duration
	keepFor  time.Duration
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Storage interface {
	EventsToNotify(ctx context.Context, now time.Time) ([]storage.Event, error)
	MarkNotified(ctx context.Context, id int64) error
	DeleteOldEvents(ctx context.Context, olderThan time.Time) (int64, error)
}

func New(logger Logger, storage Storage, producer broker.Producer, interval, keepFor time.Duration) *Scheduler {
	return &Scheduler{
		logger:   logger,
		storage:  storage,
		producer: producer,
		interval: interval,
		keepFor:  keepFor,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.notify(ctx)
	s.cleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("планировщик остановлен")
			return
		case <-ticker.C:
			s.notify(ctx)
			s.cleanup(ctx)
		}
	}
}

func (s *Scheduler) notify(ctx context.Context) {
	events, err := s.storage.EventsToNotify(ctx, time.Now().UTC())
	if err != nil {
		s.logger.Error("планировщик: получение событий для уведомлений: " + err.Error())
		return
	}

	for _, event := range events {
		notification := storage.Notification{
			EventID: event.ID,
			Title:   event.Title,
			Date:    event.StartAt,
			UserID:  event.UserID,
		}

		payload, err := json.Marshal(notification)
		if err != nil {
			s.logger.Error("планировщик: сериализация уведомления: " + err.Error())
			continue
		}

		message := broker.Message{
			Key:   []byte(strconv.FormatInt(event.ID, 10)),
			Value: payload,
		}

		if err := s.producer.Send(ctx, message); err != nil {
			s.logger.Error("планировщик: отправка уведомления: " + err.Error())
			continue
		}

		if err := s.storage.MarkNotified(ctx, event.ID); err != nil {
			s.logger.Error("планировщик: отметка события уведомлённым: " + err.Error())
			continue
		}

		s.logger.Info("планировщик: отправлено уведомление о событии " + strconv.FormatInt(event.ID, 10))
	}
}

func (s *Scheduler) cleanup(ctx context.Context) {
	threshold := time.Now().UTC().Add(-s.keepFor)

	deleted, err := s.storage.DeleteOldEvents(ctx, threshold)
	if err != nil {
		s.logger.Error("планировщик: удаление старых событий: " + err.Error())
		return
	}

	if deleted > 0 {
		s.logger.Info("планировщик: удалено старых событий: " + strconv.FormatInt(deleted, 10))
	}
}
