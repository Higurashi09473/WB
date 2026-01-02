package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func MustProducer(brokers []string, topic string) *Producer {
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
		log.Fatalf("%s: failed to dial kafka broker %s: %v", op, brokers[0], err)
	}
	defer conn.Close()

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
