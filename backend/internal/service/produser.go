package service

import (
    "context"
    "fmt"
    "time"

    "github.com/segmentio/kafka-go"
)

type Producer struct {
    writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
    writer := &kafka.Writer{
        Addr:                   kafka.TCP(brokers...),
        Topic:                  topic,
        Balancer:               &kafka.LeastBytes{},
        WriteTimeout:           5 * time.Second,
        RequiredAcks:           kafka.RequireOne,
        AllowAutoTopicCreation: true,
    }

    return &Producer{
        writer: writer,
    }, nil
}

func (p *Producer) Send(ctx context.Context, key string, value []byte) error {
    err := p.writer.WriteMessages(ctx, kafka.Message{
        Key:   []byte(key),
        Value: value,
    })
    if err != nil {
        return fmt.Errorf("failed to send message: %w", err)
    }
    return nil
}

func (p *Producer) Close() {
    p.writer.Close()
}