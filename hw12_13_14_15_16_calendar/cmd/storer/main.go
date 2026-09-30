package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/broker/kafka"
	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/config"
	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/logger"
	sqlstorage "github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storage/sql"
	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storer"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Путь к файлу конфигурации")
}

func main() {
	flag.Parse()

	cfg, err := config.NewConfig(configFile)
	if err != nil {
		panic(err)
	}

	logg := logger.New(cfg.Logger.Level)

	store, err := sqlstorage.New(cfg.Storage.DSN, logg)
	if err != nil {
		logg.Error("подключение к БД: " + err.Error())
		os.Exit(1)
	}
	defer func() { _ = store.Close() }()

	consumer, err := kafka.NewConsumer(kafka.Config{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
		GroupID: cfg.Kafka.GroupID,
	}, logg)
	if err != nil {
		logg.Error("подключение к Kafka: " + err.Error())
		os.Exit(1)
	}
	defer func() { _ = consumer.Close() }()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	storer.New(logg, store, consumer).Run(ctx)
}
