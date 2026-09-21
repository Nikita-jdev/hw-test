package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/fixme_my_friend/hw12_13_14_15_16_calendar/internal/broker"
)

const (
	retryTimeout = time.Second
	retryMax     = 10
)

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

type Producer struct {
	writer *kafkago.Writer
	logger Logger
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

func NewProducer(cfg Config, logger Logger) (broker.Producer, error) {
	if len(cfg.Brokers) == 0 || cfg.Topic == "" {
		return nil, errors.New("kafka: не заданы брокеры или топик")
	}

	return &Producer{
		writer: &kafkago.Writer{
			Addr:                   kafkago.TCP(cfg.Brokers...),
			Topic:                  cfg.Topic,
			Balancer:               &kafkago.LeastBytes{},
			RequiredAcks:           kafkago.RequireOne,
			AllowAutoTopicCreation: true,
		},
		logger: logger,
	}, nil
}

func (p *Producer) Send(ctx context.Context, message broker.Message) error {
	var err error

	for attempt := 1; attempt <= retryMax; attempt++ {
		err = p.writer.WriteMessages(ctx, kafkago.Message{Key: message.Key, Value: message.Value})
		if err == nil {
			return nil
		}

		p.logger.Error(fmt.Sprintf("kafka: ошибка отправки сообщения (попытка %d): %s", attempt, err))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryTimeout):
		}
	}
	return fmt.Errorf("kafka: не удалось отправить сообщение: %w", err)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

type Consumer struct {
	reader *kafkago.Reader
	logger Logger
}

func NewConsumer(cfg Config, logger Logger) (broker.Consumer, error) {
	if len(cfg.Brokers) == 0 || cfg.Topic == "" {
		return nil, errors.New("kafka: не заданы брокеры или топик")
	}
	if cfg.GroupID == "" {
		return nil, errors.New("kafka: не задан group_id")
	}

	return &Consumer{
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:        cfg.Brokers,
			GroupID:        cfg.GroupID,
			Topic:          cfg.Topic,
			MinBytes:       10,   // 10B
			MaxBytes:       10e6, // 10MB
			CommitInterval: 0,    // подтверждать вручную через Commit
			StartOffset:    kafkago.LastOffset,
		}),
		logger: logger,
	}, nil
}

func (c *Consumer) Receive(ctx context.Context) (broker.Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return broker.Message{}, err
	}

	return broker.Message{Key: msg.Key, Value: msg.Value}, nil
}

func (c *Consumer) Commit(ctx context.Context) error {
	return c.reader.CommitMessages(ctx)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
