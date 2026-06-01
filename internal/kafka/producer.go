package kafka

import (
	"context"
	"encoding/json"

	kafkago "github.com/segmentio/kafka-go"
)

// Producer публикует JSON-сообщения в один Kafka-топик.
type Producer struct {
	writer *kafkago.Writer
	topic  string
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		topic: topic,
		writer: &kafkago.Writer{
			Addr:                   kafkago.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafkago.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *Producer) Topic() string { return p.topic }

// Publish сериализует value в JSON и отправляет в топик с данным ключом.
func (p *Producer) Publish(ctx context.Context, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(key),
		Value: data,
	})
}

func (p *Producer) Close() error { return p.writer.Close() }
