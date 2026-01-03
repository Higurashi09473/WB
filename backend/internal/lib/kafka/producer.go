package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func MustProducer(log *slog.Logger, brokers []string, topic string) *Producer {
	const op = "kafka.produser.MustProducer"

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
		log.Error("failed to dial kafka broker",
			slog.String("op", op),
			slog.String("broker", brokers[0]),
			slog.Any("err", err))
		os.Exit(1)
	}

	if err := conn.Close(); err != nil {
		log.Error("failed close connection",
			slog.String("op", op),
			slog.String("broker", brokers[0]),
			slog.Any("err", err))
		os.Exit(1)
	}

	return &Producer{writer: writer}
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

func (p *Producer) Close() error {
	return p.writer.Close()
}
