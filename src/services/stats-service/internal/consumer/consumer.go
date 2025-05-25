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
	Timestamp   time.Time `json:"viewed_at,omitempty"`
	LikedAt     time.Time `json:"liked_at,omitempty"`
	UnlikedAt   time.Time `json:"unliked_at,omitempty"`
	CommentedAt time.Time `json:"created_at,omitempty"`
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

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("reader error %s: %v", topic, err)
			time.Sleep(time.Second)
			continue
		}
		var ev Event
		if err := json.Unmarshal(m.Value, &ev); err != nil {
			log.Printf("unmarshal %s: %v", topic, err)
			continue
		}
		// Определяем дату и метрику
		date := ev.Timestamp.UTC().Format("2006-01-02")
		metric := topic // "post-views" и т.п.
		// Вставляем в ClickHouse
		err = ck.Exec(context.Background(),
			`INSERT INTO stats.events (metric, entity_id, event_date, cnt) VALUES (?, ?, ?, 1)`,
			metric, ev.PostID, date,
		)
		if err != nil {
			log.Printf("clickhouse insert %s: %v", topic, err)
		}
	}
}
