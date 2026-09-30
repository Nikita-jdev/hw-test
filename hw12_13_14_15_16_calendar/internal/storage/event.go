package storage

import (
	"errors"
	"time"
)

type Event struct {
	ID           int64
	Title        string
	StartAt      time.Time     // дата и время события
	Duration     time.Duration // длительность события
	UserID       int64         // ID владельца события
	NotifyBefore time.Duration // за сколько до события напомнить (0 — не напоминать)
	Notified     bool
}

type Notification struct {
	EventID int64     `json:"eventId"`
	Title   string    `json:"title"`
	Date    time.Time `json:"date"`
	UserID  int64     `json:"userId"`
}

var (
	ErrDateBusy      = errors.New("данное время уже занято другим событием")
	ErrEventNotFound = errors.New("событие не найдено")
)
