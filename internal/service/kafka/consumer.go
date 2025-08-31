package kafka

import (
    "context"

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

// Consume читает одно сообщение из Kafka
func (c *Consumer) Consume(ctx context.Context) (string, []byte, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return "", nil, err
	}
	return string(msg.Key), msg.Value, nil
}

func (c *Consumer) Close() {
    c.reader.Close()
}