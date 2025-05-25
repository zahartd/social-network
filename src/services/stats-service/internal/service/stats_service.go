package service

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type StatsService struct{ ck clickhouse.Conn }

func NewStatsService(ck clickhouse.Conn) *StatsService {
	return &StatsService{ck: ck}
}

func (s *StatsService) GetPostStats(postID string) (views, likes, comments uint64, err error) {
	rows, err := s.ck.Query(context.Background(),
		`SELECT metric, sum(cnt) FROM stats.events WHERE entity_id = ? GROUP BY metric`,
		postID,
	)
	if err != nil {
		return
	}
	defer rows.Close()
	var metric string
	var sum uint64
	for rows.Next() {
		rows.Scan(&metric, &sum)
		switch metric {
		case "post-views":
			views = sum
		case "post-likes":
			likes = sum
		case "post-comments":
			comments = sum
		case "post-unlikes": /* вычитаем или не учитываем? */
		}
	}
	return
}

func (s *StatsService) GetDynamics(postID string, metric string) ([]struct {
	Date  string
	Count uint64
}, error) {
	rows, err := s.ck.Query(context.Background(),
		`SELECT event_date, sum(cnt) FROM stats.events
        WHERE entity_id = ? AND metric = ? GROUP BY event_date ORDER BY event_date`,
		postID, metric,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		Date  string
		Count uint64
	}
	for rows.Next() {
		var d time.Time
		var c uint64
		rows.Scan(&d, &c)
		out = append(out, struct {
			Date  string
			Count uint64
		}{Date: d.Format("2006-01-02"), Count: c})
	}
	return out, nil
}

func (s *StatsService) GetTop(by string) ([]struct {
	ID    string
	Count uint64
}, error) {
	rows, err := s.ck.Query(context.Background(),
		`SELECT entity_id, sum(cnt) as c FROM stats.events WHERE metric = ?
        GROUP BY entity_id ORDER BY c DESC LIMIT 10`,
		by,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		ID    string
		Count uint64
	}
	for rows.Next() {
		var id string
		var c uint64
		rows.Scan(&id, &c)
		out = append(out, struct {
			ID    string
			Count uint64
		}{ID: id, Count: c})
	}
	return out, nil
}
