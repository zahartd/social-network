package service

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type DayCount struct {
	Date  string
	Count uint64
}
type TopItem struct {
	ID    string
	Count uint64
}

type StatsService struct {
	ck clickhouse.Conn
}

func NewStatsService(ck clickhouse.Conn) *StatsService {
	return &StatsService{ck: ck}
}

func (s *StatsService) GetPostStats(postID string) (views, likes, comments uint64, err error) {
	const q = `
        SELECT metric, sum(cnt)
          FROM stats.events
         WHERE entity_id = ?
         GROUP BY metric`
	rows, err := s.ck.Query(context.Background(), q, postID)
	if err != nil {
		return
	}
	defer rows.Close()

	var metric string
	var sumCnt uint64
	for rows.Next() {
		if err = rows.Scan(&metric, &sumCnt); err != nil {
			return
		}
		switch metric {
		case "post-views":
			views = sumCnt
		case "post-likes":
			likes = sumCnt
		case "post-unlikes":
			// subtract
			if sumCnt > likes {
				likes = 0
			} else {
				likes -= sumCnt
			}
		case "post-comments":
			comments = sumCnt
		}
	}
	return
}

func (s *StatsService) GetDynamics(postID, metric string) ([]DayCount, error) {
	const q = `
        SELECT event_date, sum(cnt)
          FROM stats.events
         WHERE entity_id = ? AND metric = ?
         GROUP BY event_date
         ORDER BY event_date`
	rows, err := s.ck.Query(context.Background(), q, postID, metric)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DayCount
	for rows.Next() {
		var d time.Time
		var c uint64
		if err := rows.Scan(&d, &c); err != nil {
			return nil, err
		}
		out = append(out, DayCount{Date: d.Format("2006-01-02"), Count: c})
	}
	return out, nil
}

func (s *StatsService) GetTopPosts(metric string) ([]TopItem, error) {
	const q = `
        SELECT entity_id, sum(cnt) as c
          FROM stats.events
         WHERE metric = ?
         GROUP BY entity_id
         ORDER BY c DESC
         LIMIT 10`
	rows, err := s.ck.Query(context.Background(), q, metric)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TopItem
	for rows.Next() {
		var id string
		var c uint64
		if err := rows.Scan(&id, &c); err != nil {
			return nil, err
		}
		out = append(out, TopItem{ID: id, Count: c})
	}
	return out, nil
}

func (s *StatsService) GetTopUsers(metric string) ([]TopItem, error) {
	const q = `
        SELECT user_id, sum(cnt) as c
          FROM stats.events
         WHERE metric = ?
         GROUP BY user_id
         ORDER BY c DESC
         LIMIT 10`
	rows, err := s.ck.Query(context.Background(), q, metric)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TopItem
	for rows.Next() {
		var id string
		var c uint64
		if err := rows.Scan(&id, &c); err != nil {
			return nil, err
		}
		out = append(out, TopItem{ID: id, Count: c})
	}
	return out, nil
}
