package logger

import (
	"fmt"
	"strings"
)

type Logger struct {
	level int
}

func New(level string) *Logger {
	levels := map[string]int{"debug": 0, "info": 1, "warn": 2, "error": 3}
	return &Logger{level: levels[strings.ToLower(level)]}
}

func (l Logger) Debug(msg string) {
	l.log(0, "Отладка: ", msg)
}

func (l Logger) Info(msg string) {
	l.log(1, "Информация: ", msg)
}

func (l Logger) Warn(msg string) {
	l.log(2, "Предупреждение: ", msg)
}

func (l Logger) Error(msg string) {
	l.log(3, "Ошибка: ", msg)
}

func (l Logger) log(level int, prefix, msg string) {
	if level >= l.level {
		fmt.Println(prefix + msg)
	}
}
