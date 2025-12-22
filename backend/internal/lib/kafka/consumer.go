package kafka

import (
	"context"
	"errors"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
    reader *kafka.Reader
}

func NewConsumer(brokers []string, group, topic string) *Consumer {
    return &Consumer{
        reader: kafka.NewReader(kafka.ReaderConfig{
            Brokers:  brokers,
            GroupID:  group,
            Topic:    topic,
            MinBytes: 10e3, // 10KB
            MaxBytes: 10e6, // 10MB
        }),
    }
}

type MessageHandler func(ctx context.Context, key string, value []byte) error

func (c *Consumer) Start(ctx context.Context, handler MessageHandler) error {
    for {
        msg, err := c.reader.FetchMessage(ctx)
        if err != nil {
            if errors.Is(err, context.Canceled){
                return nil
            }
            return err
        }

        if err := handler(ctx, string(msg.Key), msg.Value); err != nil {
            continue
        }

        if err := c.reader.CommitMessages(ctx, msg); err != nil {
            // Нужно добавить retry или DLQ
        }
    }
}

func (c *Consumer) Close() error {
    return c.reader.Close()
}