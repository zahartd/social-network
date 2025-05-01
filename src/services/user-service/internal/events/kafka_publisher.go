package events

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"os"

	"github.com/hamba/avro/registry"
	"github.com/segmentio/kafka-go"
)

//go:generate go run github.com/hamba/avro/v2/cmd/avrogen -pkg events -encoders -fullschema -o user_registered_gen.go schemas/user_registered.avsc

type KafkaPublisher interface {
	WriteUserRegistered(ctx context.Context, ev UserRegistered) error
}

type kafkaPublisherJSON struct {
	writer *kafka.Writer
}

func NewKafkaPublisherJSON(brokers []string) (KafkaPublisher, error) {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   "user-registrations",
		Async:   true,
	})
	return &kafkaPublisherJSON{writer: w}, nil
}

func (p *kafkaPublisherJSON) WriteUserRegistered(ctx context.Context, ev UserRegistered) error {
	buf, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	msg := kafka.Message{Key: []byte(ev.UserID), Value: buf}
	return p.writer.WriteMessages(ctx, msg)
}

type kafkaPublisherAvro struct {
	writer   *kafka.Writer
	registry *registry.Client
	schemaID int
}

func NewKafkaPublisherAvro(brokers []string, regURL string) (KafkaPublisher, error) {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER_URL")},
		Topic:   "user-registrations",
		Async:   true,
	})

	reg, err := registry.NewClient(regURL)
	if err != nil {
		return nil, err
	}

	id, _, err := reg.CreateSchema("users.UserRegistered-value", schemaUserRegistered.String())
	if err != nil {
		return nil, err
	}

	return &kafkaPublisherAvro{writer: w, registry: reg, schemaID: id}, nil
}

func (p *kafkaPublisherAvro) WriteUserRegistered(ctx context.Context, ev UserRegistered) error {
	payload, err := ev.Marshal()
	if err != nil {
		return err
	}

	wire := make([]byte, 5+len(payload))
	wire[0] = 0
	binary.BigEndian.PutUint32(wire[1:], uint32(p.schemaID))
	copy(wire[5:], payload)

	return p.writer.WriteMessages(ctx,
		kafka.Message{Key: []byte(ev.UserID), Value: wire})
}
