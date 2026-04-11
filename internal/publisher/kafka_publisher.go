package publisher

import (
	"aurum/internal/domain"
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writers map[string]*kafka.Writer
	brokers []string
}

func NewKafkaPublisher(brokers []string) *KafkaPublisher {
	return &KafkaPublisher{
		writers: make(map[string]*kafka.Writer),
		brokers: brokers,
	}
}

func (p *KafkaPublisher) Publish(ctx context.Context, event domain.OutboxEvent) error {
	writer := p.writerForTopic(event.EventType)

	msg := kafka.Message{
		Key:   []byte(event.AggregateID.String()),
		Value: event.Payload,
		Headers: []kafka.Header{
			{Key: "event_id", Value: []byte(event.ID.String())},
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "occurred_at", Value: []byte(event.CreatedAt.Format(time.RFC3339))},
		},
	}

	if err := writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message to topic %s: %w", event.EventType, err)
	}

	return nil
}

func (p *KafkaPublisher) writerForTopic(topic string) *kafka.Writer {
	if w, ok := p.writers[topic]; ok {
		return w
	}
	w := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}
	p.writers[topic] = w
	return w
}
