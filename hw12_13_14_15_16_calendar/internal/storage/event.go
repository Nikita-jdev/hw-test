package storage

import (
	"errors"
	"time"
)

type Event struct {
	ID       int64
	Title    string
	StartAt  time.Time     // дата и время события
	Duration time.Duration // длительность события
	UserID   int64         // ID владельца события
}

var (
	ErrDateBusy      = errors.New("данное время уже занято другим событием")
	ErrEventNotFound = errors.New("событие не найдено")
)
