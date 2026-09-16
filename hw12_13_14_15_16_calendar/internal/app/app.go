package app

import (
	"context"

	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storage"
)

type App struct {
	logger  Logger
	storage Storage
}

type Logger interface {
	Info(msg string)
}

type Storage interface {
	CreateEvent(ctx context.Context, event storage.Event) error
}

func New(logger Logger, storage Storage) *App {
	return &App{logger: logger, storage: storage}
}

func (a *App) CreateEvent(ctx context.Context, id int64, title string) error {
	return a.storage.CreateEvent(ctx, storage.Event{ID: id, Title: title})
}
