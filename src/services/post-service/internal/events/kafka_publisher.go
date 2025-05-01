package events

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"log"

	"github.com/hamba/avro/registry"
	"github.com/segmentio/kafka-go"
	"github.com/zahartd/social-network/src/services/post-service/internal/formats"
)

//go:generate go run github.com/hamba/avro/v2/cmd/avrogen -pkg events -encoders -fullschema -o added_comment_gen.go schemas/added_comment.avsc
//go:generate go run github.com/hamba/avro/v2/cmd/avrogen -pkg events -encoders -fullschema -o liked_post_gen.go schemas/liked_post.avsc
//go:generate go run github.com/hamba/avro/v2/cmd/avrogen -pkg events -encoders -fullschema -o viewed_post.gen.go schemas/viewed_post.avsc

type KafkaPublisher interface {
	WriteAddedComment(ctx context.Context, ev AddedCommentEvent) error
	WriteLikedPost(ctx context.Context, ev LikedPostEvent) error
	WriteViewedPost(ctx context.Context, ev ViewedPostEvent) error
}

type kafkaPublisherJSON struct {
	writers map[string]*kafka.Writer
}

func NewKafkaPublisherJSON() *kafkaPublisherJSON {
	return &kafkaPublisherJSON{
		writers: make(map[string]*kafka.Writer, 0),
	}
}

func WithWriter(pub *kafkaPublisherJSON, writer *kafka.Writer) *kafkaPublisherJSON {
	pub.writers[writer.Topic] = writer
	return pub
}

func (p *kafkaPublisherJSON) writeEvents(ctx context.Context, topic, key string, ev any) error {
	buf, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	msg := kafka.Message{Key: []byte(key), Value: buf}
	return p.writers[topic].WriteMessages(ctx, msg)
}

func (p *kafkaPublisherJSON) WriteAddedComment(ctx context.Context, ev AddedCommentEvent) error {
	return p.writeEvents(ctx, "post-comments", ev.UserID, ev)
}

func (p *kafkaPublisherJSON) WriteLikedPost(ctx context.Context, ev LikedPostEvent) error {
	return p.writeEvents(ctx, "post-likes", ev.UserID, ev)
}

func (p *kafkaPublisherJSON) WriteViewedPost(ctx context.Context, ev ViewedPostEvent) error {
	return p.writeEvents(ctx, "post-views", ev.UserID, ev)
}

type kafkaPublisherAvro struct {
	writers   map[string]*kafka.Writer
	registry  *registry.Client
	schemasID map[string]int
}

func NewKafkaPublisherAvro(brokers []string, regURL string) (KafkaPublisher, error) {
	reg, err := registry.NewClient(regURL)
	if err != nil {
		return nil, err
	}

	pub := &kafkaPublisherAvro{
		registry:  reg,
		writers:   make(map[string]*kafka.Writer, 0),
		schemasID: make(map[string]int, 0),
	}
	return pub, nil
}

func (p *kafkaPublisherAvro) WithAddCommentWriter(writer *kafka.Writer) KafkaPublisher {
	p.writers[writer.Topic] = writer
	id, _, err := p.registry.CreateSchema("posts.AddedComment-value", schemaAddedCommentEvent.String())
	if err != nil {
		log.Fatalf("failed to register schema to topic %s: %s", writer.Topic, err.Error())
	}

	p.writers[writer.Topic] = writer
	p.schemasID[writer.Topic] = id

	return p
}

func (p *kafkaPublisherAvro) WithLikePostWriter(writer *kafka.Writer) KafkaPublisher {
	p.writers[writer.Topic] = writer
	id, _, err := p.registry.CreateSchema("posts.LikedPost-value", schemaLikedPostEvent.String())
	if err != nil {
		log.Fatalf("failed to register schema to topic %s: %s", writer.Topic, err.Error())
	}

	p.writers[writer.Topic] = writer
	p.schemasID[writer.Topic] = id

	return p
}

func (p *kafkaPublisherAvro) WithViewPostWriter(writer *kafka.Writer) KafkaPublisher {
	p.writers[writer.Topic] = writer
	id, _, err := p.registry.CreateSchema("posts.ViewedPost-value", schemaViewedPostEvent.String())
	if err != nil {
		log.Fatalf("failed to register schema to topic %s: %s", writer.Topic, err.Error())
	}

	p.writers[writer.Topic] = writer
	p.schemasID[writer.Topic] = id

	return p
}

func (p *kafkaPublisherAvro) writeEvent(ctx context.Context, topic, key string, ev formats.Marshaler) error {
	payload, err := ev.Marshal()
	if err != nil {
		return err
	}

	wire := make([]byte, 5+len(payload))
	wire[0] = 0
	binary.BigEndian.PutUint32(wire[1:], uint32(p.schemasID[topic]))
	copy(wire[5:], payload)

	return p.writers[topic].WriteMessages(ctx,
		kafka.Message{Key: []byte(key), Value: wire})
}

func (p *kafkaPublisherAvro) WriteAddedComment(ctx context.Context, ev AddedCommentEvent) error {
	return p.writeEvent(ctx, "post-comments", ev.UserID, &ev)
}

func (p *kafkaPublisherAvro) WriteLikedPost(ctx context.Context, ev LikedPostEvent) error {
	return p.writeEvent(ctx, "post-likes", ev.UserID, &ev)
}

func (p *kafkaPublisherAvro) WriteViewedPost(ctx context.Context, ev ViewedPostEvent) error {
	return p.writeEvent(ctx, "post-views", ev.UserID, &ev)
}
