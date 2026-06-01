package kafka

import (
	"context"
	"encoding/json"

	kafkago "github.com/segmentio/kafka-go"
)

// Consumer читает JSON-сообщения из одного Kafka-топика (раздел 0).
// Начинает с LastOffset — обрабатывает только новые сообщения.
type Consumer struct {
	reader *kafkago.Reader
	topic  string
}

func NewConsumer(brokers []string, topic string) *Consumer {
	return &Consumer{
		topic: topic,
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:     brokers,
			Topic:       topic,
			Partition:   0,
			StartOffset: kafkago.LastOffset,
		}),
	}
}

func (c *Consumer) Topic() string { return c.topic }

// Read блокируется до получения следующего сообщения, десериализует его в dest.
func (c *Consumer) Read(ctx context.Context, dest any) error {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return err
	}
	return json.Unmarshal(msg.Value, dest)
}

func (c *Consumer) Close() error { return c.reader.Close() }
