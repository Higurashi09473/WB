package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)


type dlqMessage struct {
	OriginalTopic     string            `json:"original_topic"`
	OriginalPartition int               `json:"original_partition"`
	OriginalOffset    int64             `json:"original_offset"`
	OriginalKey       string            `json:"original_key"`
	OriginalValue     json.RawMessage   `json:"original_value"`
	Error             string            `json:"error"`
	FailedAt          time.Time         `json:"failed_at"`
	Headers           map[string]string `json:"headers,omitempty"`
}

// sendToDLQ sends the original message to the dead letter queue with error context.
func (c *Consumer) sendToDLQ(ctx context.Context, msg kafka.Message, procErr error) error {
	const op = "kafka.dlq.sendToDLQ"

    if c.dlqWriter == nil {
		return fmt.Errorf("%s: errors during close: %v", op, "dlq writer not configured")
	}

	headers := make(map[string]string)
	for _, h := range msg.Headers {
		headers[h.Key] = string(h.Value)
	}

	dlqMsg := dlqMessage{
		OriginalTopic:     msg.Topic,
		OriginalPartition: msg.Partition,
		OriginalOffset:    msg.Offset,
		OriginalKey:       string(msg.Key),
		OriginalValue:     msg.Value,
		Error:             procErr.Error(),
		FailedAt:          time.Now().UTC(),
		Headers:           headers,
	}

	valueBytes, err := json.Marshal(dlqMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal dlq message: %w", err)
	}

	return c.dlqWriter.WriteMessages(ctx, kafka.Message{
		Key:   msg.Key,
		Value: valueBytes,
		Headers: []kafka.Header{
			{Key: "dlq_reason", Value: []byte(procErr.Error())},
			{Key: "original_topic", Value: []byte(msg.Topic)},
		},
	})
}