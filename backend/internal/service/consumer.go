package service

import (
    "context"
    "log"

    "github.com/segmentio/kafka-go"
)

type Consumer struct {
    reader *kafka.Reader
}

func NewConsumer(brokers []string, group, topic string) (*Consumer, error) {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers:  brokers,
        GroupID:  group,
        Topic:    topic,
        MinBytes: 10e3, // 10KB
        MaxBytes: 10e6, // 10MB
    })

    return &Consumer{reader: reader}, nil
}

func (c *Consumer) Run(ctx context.Context, handler func(msg *kafka.Message) error) {
    for {
        msg, err := c.reader.FetchMessage(ctx)
        if err != nil {
            log.Printf("kafka fetch error: %v", err)
            continue
        }

        if err := handler(&msg); err == nil {
            if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
                log.Printf("commit error: %v", commitErr)
            }
        } else {
            log.Printf("handler error: %v", err)
        }
    }
}

func (c *Consumer) Close() {
    c.reader.Close()
}