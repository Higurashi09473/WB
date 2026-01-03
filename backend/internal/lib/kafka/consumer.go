// Package kafka provides Kafka producer and consumer implementations
// for the message broker interface used in the WB backend
package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)
// Consumer represents Message broker consumer.
type Consumer struct {
	reader    *kafka.Reader
	dlqWriter *kafka.Writer
}

// MessageHandler is a function type for processing incoming Kafka messages.
type MessageHandler func(ctx context.Context, value []byte) error


// NewConsumer creates and configures a new Kafka consumer with DLQ writer.
// It connects to the specified brokers, joins the consumer group and subscribes to the topic.
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

// Start begins consuming messages from Kafka and processes them using the provided handler.
// It runs until the context is canceled or a fatal error occurs.
// On handler error — message is skipped (not committed), but consumption continues.
// On commit error — message is sent to DLQ if possible.
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

		if err := handler(ctx, msg.Value); err != nil {
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			_ = c.sendToDLQ(ctx, msg, err)
		}
	}
}

// Close gracefully shuts down the consumer and DLQ writer.
// It closes both connections and collects any errors.
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
