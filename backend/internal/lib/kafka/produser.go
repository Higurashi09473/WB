package kafka

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
    const op = "kafka.produser.NewProducer"

    writer := &kafka.Writer{
        Addr:                   kafka.TCP(brokers...),
        Topic:                  topic,
        Balancer:               &kafka.LeastBytes{},
        WriteTimeout:           5 * time.Second,
        RequiredAcks:           kafka.RequireOne,
        AllowAutoTopicCreation: true,
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
    if err != nil {
        return nil, fmt.Errorf("%s: failed to dial kafka broker %s: %w", op, brokers[0], err)
    }
    defer conn.Close()

    return &Producer{writer: writer}, nil
}

func (p *Producer) Send(ctx context.Context, key string, value []byte) error {
    const op = "kafka.produser.Send"

    err := p.writer.WriteMessages(ctx, kafka.Message{
        Key:   []byte(key),
        Value: value,
    })
    if err != nil {
        return fmt.Errorf("%s: failed to send message: %w", op, err)
    }
    return nil
}

func (p *Producer) Close() error{
    return p.writer.Close()
}