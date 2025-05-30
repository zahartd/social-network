package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/segmentio/kafka-go"
	"github.com/zahartd/social-network/src/services/stats-service/internal/config"
)

type Event struct {
	UserID      string    `json:"user_id"`
	PostID      string    `json:"post_id"`
	ViewedAt    time.Time `json:"viewed_at"`
	LikedAt     time.Time `json:"liked_at"`
	UnlikedAt   time.Time `json:"unliked_at"`
	CommentedAt time.Time `json:"created_at"`
}

func RunAll(cfg *config.Config, ck clickhouse.Conn) {
	topics := []string{"post-views", "post-likes", "post-unlikes", "post-comments"}
	for _, topic := range topics {
		go runConsumer(cfg, ck, topic)
	}
}

func runConsumer(cfg *config.Config, ck clickhouse.Conn, topic string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KafkaBrokerURL},
		Topic:   topic,
		GroupID: "stats-service",
	})
	defer r.Close()

	ctx := context.Background()
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("[Consumer %s] read error: %v", topic, err)
			time.Sleep(time.Second)
			continue
		}

		var ev Event
		if err := json.Unmarshal(m.Value, &ev); err != nil {
			log.Printf("[Consumer %s] json unmarshal error: %v", topic, err)
			continue
		}

		var ts time.Time
		switch topic {
		case "post-views":
			ts = ev.ViewedAt
		case "post-likes":
			ts = ev.LikedAt
		case "post-unlikes":
			ts = ev.UnlikedAt
		case "post-comments":
			ts = ev.CommentedAt
		default:
			continue
		}

		const q = `
            INSERT INTO stats.events
                (event_time, user_id, entity_id, metric, cnt)
            VALUES (?, ?, ?, ?, ?)`
		if err := ck.Exec(ctx, q,
			ts,
			ev.UserID,
			ev.PostID,
			topic,
			uint64(1),
		); err != nil {
			log.Printf("[Consumer %s] clickhouse insert error: %v", topic, err)
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			log.Printf("[Consumer %s] commit offset error: %v", topic, err)
		}
	}
}
