package storer

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/broker"
	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storage"
)

type Storer struct {
	logger   Logger
	storage  Storage
	consumer broker.Consumer
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Storage interface {
	SaveNotification(ctx context.Context, notification storage.Notification) error
}

func New(logger Logger, storage Storage, consumer broker.Consumer) *Storer {
	return &Storer{
		logger:   logger,
		storage:  storage,
		consumer: consumer,
	}
}

func (s *Storer) Run(ctx context.Context) {
	s.logger.Info("хранитель запущен")

	for {
		message, err := s.consumer.Receive(ctx)
		if err != nil {
			if ctx.Err() != nil {
				s.logger.Info("хранитель остановлен")
				return
			}
			s.logger.Error("хранитель: получение сообщения: " + err.Error())
			continue
		}

		var notification storage.Notification
		if err := json.Unmarshal(message.Value, &notification); err != nil {
			s.logger.Error("хранитель: десериализация уведомления: " + err.Error())
			continue
		}

		if err := s.storage.SaveNotification(ctx, notification); err != nil {
			s.logger.Error("хранитель: сохранение уведомления: " + err.Error())
			continue
		}

		if err := s.consumer.Commit(ctx); err != nil {
			s.logger.Error("хранитель: подтверждение сообщения: " + err.Error())
			continue
		}

		s.logger.Info("хранитель: сохранено уведомление о событии " + strconv.FormatInt(notification.EventID, 10))
	}
}
