package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/broker/kafka"
	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/config"
	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/scheduler"
	sqlstorage "github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storage/sql"
)

const (
	defaultInterval = time.Second * 30
	defaultKeepFor  = time.Hour * 24 * 365
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

	producer, err := kafka.NewProducer(kafka.Config{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
	}, logg)
	if err != nil {
		logg.Error("подключение к Kafka: " + err.Error())
		os.Exit(1)
	}
	defer func() { _ = producer.Close() }()

	interval := time.Duration(cfg.Scheduler.Interval)
	if interval <= 0 {
		interval = defaultInterval
	}

	keepFor := time.Duration(cfg.Scheduler.KeepPeriod)
	if keepFor <= 0 {
		keepFor = defaultKeepFor
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	logg.Info("планировщик запущен")

	scheduler.New(logg, store, producer, interval, keepFor).Run(ctx)
}
