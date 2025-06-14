package kafka_consumer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/zahartd/social-network/src/services/stats-service/internal/domain/models"
	"github.com/zahartd/social-network/src/services/stats-service/internal/domain/repository"
)

type Consumer struct {
	readers []*kafka.Reader
	repo    repository.Stats
}

func New(broker string, repo repository.Stats) *Consumer {
	topics := []string{"post-views", "post-likes", "post-unlikes", "post-comments"}
	var rs []*kafka.Reader
	for _, t := range topics {
		rs = append(rs, kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{broker},
			Topic:   t,
			GroupID: "stats-service",
			// ключевая строка: начинаем с самого раннего, если нет offset’а
			StartOffset:    kafka.FirstOffset,
			MinBytes:       1,
			MaxBytes:       1 << 20,
			CommitInterval: 0, // синхронный commit — быстрее тесты
		}))
	}
	return &Consumer{readers: rs, repo: repo}
}

func (c *Consumer) RunAsync() {
	for _, rd := range c.readers {
		go func(r *kafka.Reader) {
			ctx := context.Background()
			for {
				m, err := r.ReadMessage(ctx)
				if err != nil {
					time.Sleep(time.Second)
					continue
				}

				var raw struct {
					UserID string `json:"user_id"`
					PostID string `json:"post_id"`
				}
				if err := json.Unmarshal(m.Value, &raw); err != nil {
					continue
				}
				ev := models.Event{
					Metric: string(m.Topic),
					Time:   time.Now().UTC(),
					UserID: raw.UserID,
					PostID: raw.PostID,
					Count:  1,
				}
				_ = c.repo.Insert(ctx, ev)
				_ = r.CommitMessages(ctx, m)
			}
		}(rd)
	}
}

func (c *Consumer) Close() {
	for _, r := range c.readers {
		_ = r.Close()
	}
}
