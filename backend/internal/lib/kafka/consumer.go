package kafka

import (
    "context"

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
            if err == context.Canceled {
                // log.Println("Consumer остановлен по контексту (graceful shutdown)")
                return nil
            }
            return err
        }

        // Вызываем бизнес-логику (usecase)
        if err := handler(ctx, string(msg.Key), msg.Value); err != nil {
            continue
        }

        // Только после успешной обработки — коммитим оффсет
        if err := c.reader.CommitMessages(ctx, msg); err != nil {
            // log.Printf("Ошибка коммита оффсета: %v", err)
            // Можно добавить retry или DLQ
        }

        // log.Printf("Сообщение успешно обработано и закоммичено (key=%s)", string(msg.Key))
    }
}

func (c *Consumer) Close() error {
    return c.reader.Close()
}