package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/zahartd/social-network/src/services/user-service/internal/domain/models"
)

type UserWriter struct {
	*kafka.Writer
}

func NewUserWriter(broker string) *UserWriter {
	return &UserWriter{Writer: &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        "user-registrations",
		RequiredAcks: kafka.RequireAll,
		Async:        true,
	}}
}

func (w *UserWriter) PublishUserRegistered(ctx context.Context, u models.User) error {
	payload, _ := json.Marshal(struct {
		UserID    string    `json:"user_id"`
		CreatedAt time.Time `json:"created_at"`
		Email     string    `json:"email"`
	}{u.ID, u.CreatedAt, u.Email})
	return w.WriteMessages(ctx, kafka.Message{Key: []byte(u.ID), Value: payload})
}
