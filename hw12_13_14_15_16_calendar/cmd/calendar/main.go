package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/app"
	interputConfig "github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/config"
	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/logger"
	internalHttp "github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/server/http"
	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storage"
	memorystorage "github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storage/memory"
	sqlstorage "github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := interputConfig.NewConfig(configFile)
	if err != nil {
		panic(err)
	}

	logg := logger.New(config.Logger.Level)

	var store storage.Storage
	switch config.Storage.Type {
	case "sql":
		s, err := sqlstorage.New(config.Storage.DSN, logg)
		if err != nil {
			logg.Error("подключение к БД: " + err.Error())
			os.Exit(1)
		}
		defer func(s *sqlstorage.Storage) { _ = s.Close() }(s)
		store = s
	default: // "memory"
		store = memorystorage.New()
	}

	calendar := app.New(logg, store)

	server := internalHttp.NewServer(logg, calendar, config.HTTP.Host, config.HTTP.Port)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1)
	}
}
