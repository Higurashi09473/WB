package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader    *kafka.Reader
	dlqWriter *kafka.Writer
}

func NewConsumer(brokers []string, group, topic, dlqTopic string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			GroupID:  group,
			Topic:    topic,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
		dlqWriter: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  dlqTopic,
			Balancer:               &kafka.LeastBytes{},
			WriteTimeout:           5 * time.Second,
			RequiredAcks:           kafka.RequireOne,
			AllowAutoTopicCreation: true,
		},
	}
}

type MessageHandler func(ctx context.Context, key string, value []byte) error

func (c *Consumer) Start(ctx context.Context, handler MessageHandler) error {
    const op = "kafka.consumer.Start"

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return fmt.Errorf("%s: fetch err: %w", op, err)
		}

		if err := handler(ctx, string(msg.Key), msg.Value); err != nil {
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			if err := c.sendToDLQ(ctx, msg, err); err != nil {
				continue
			}
		}
	}
}

func (c *Consumer) Close() error {
    const op = "kafka.consumer.Close"

	var errs []error
	if c.reader != nil {
		if err := c.reader.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if c.dlqWriter != nil {
		if err := c.dlqWriter.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s: errors during close: %v", op, errs)
	}
	return nil
}
