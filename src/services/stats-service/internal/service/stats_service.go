package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type DayCount struct {
	Date  string
	Count uint64
}

type TopItem struct {
	ID    string
	Count int64
}

type StatsService struct {
	ck clickhouse.Conn
}

func NewStatsService(ck clickhouse.Conn) *StatsService {
	return &StatsService{ck: ck}
}

func (s *StatsService) GetPostStats(ctx context.Context, postID string) (views, likes, comments int64, err error) {
	var (
		v   uint64
		pl  uint64
		pul uint64
		pc  uint64
	)

	if err = s.ck.QueryRow(ctx,
		`SELECT sum(cnt) FROM stats.events WHERE entity_id = ? AND metric = 'post-views'`,
		postID,
	).Scan(&v); err != nil {
		return
	}

	if err = s.ck.QueryRow(ctx,
		`SELECT sum(cnt) FROM stats.events WHERE entity_id = ? AND metric = 'post-likes'`,
		postID,
	).Scan(&pl); err != nil {
		return
	}
	if err = s.ck.QueryRow(ctx,
		`SELECT sum(cnt) FROM stats.events WHERE entity_id = ? AND metric = 'post-unlikes'`,
		postID,
	).Scan(&pul); err != nil {
		return
	}

	if err = s.ck.QueryRow(ctx,
		`SELECT sum(cnt) FROM stats.events WHERE entity_id = ? AND metric = 'post-comments'`,
		postID,
	).Scan(&pc); err != nil {
		return
	}

	views = int64(v)
	likes = int64(pl) - int64(pul)
	comments = int64(pc)
	return
}

func (s *StatsService) GetDynamics(ctx context.Context, postID, metric string) ([]DayCount, error) {
	q := `
        SELECT toDate(event_time) AS date, sum(cnt) AS count
        FROM stats.events
        WHERE entity_id = ? AND metric = ?
        GROUP BY date
        ORDER BY date
    `
	rows, err := s.ck.Query(ctx, q, postID, metric)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []DayCount
	for rows.Next() {
		var date time.Time
		var c uint64
		if err := rows.Scan(&date, &c); err != nil {
			return nil, err
		}
		res = append(res, DayCount{
			Date:  date.Format("2006-01-02"),
			Count: c,
		})
	}
	return res, nil
}

func (s *StatsService) GetTopPosts(ctx context.Context, metric string) ([]TopItem, error) {
	var q string
	if metric == "post-likes" {
		q = `
            SELECT
              entity_id,
              sumIf(cnt, metric='post-likes') - sumIf(cnt, metric='post-unlikes') AS total
            FROM stats.events
            GROUP BY entity_id
            ORDER BY total DESC
            LIMIT 10
        `
	} else {
		q = fmt.Sprintf(`
            SELECT entity_id, sum(cnt) AS total
            FROM stats.events
            WHERE metric = '%s'
            GROUP BY entity_id
            ORDER BY total DESC
            LIMIT 10
        `, metric)
	}

	rows, err := s.ck.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TopItem
	for rows.Next() {
		if metric == "post-likes" {
			var id string
			var total int64
			if err := rows.Scan(&id, &total); err != nil {
				return nil, err
			}
			out = append(out, TopItem{ID: id, Count: total})
		} else {
			var id string
			var total uint64
			if err := rows.Scan(&id, &total); err != nil {
				return nil, err
			}
			out = append(out, TopItem{ID: id, Count: int64(total)})
		}
	}
	return out, nil
}

func (s *StatsService) GetTopUsers(ctx context.Context, metric string) ([]TopItem, error) {
	var q string
	if metric == "post-likes" {
		q = `
            SELECT
              user_id,
              sumIf(cnt, metric='post-likes') - sumIf(cnt, metric='post-unlikes') AS total
            FROM stats.events
            GROUP BY user_id
            ORDER BY total DESC
            LIMIT 10
        `
	} else {
		q = fmt.Sprintf(`
            SELECT user_id, sum(cnt) AS total
            FROM stats.events
            WHERE metric = '%s'
            GROUP BY user_id
            ORDER BY total DESC
            LIMIT 10
        `, metric)
	}

	rows, err := s.ck.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TopItem
	for rows.Next() {
		if metric == "post-likes" {
			var id string
			var total int64
			if err := rows.Scan(&id, &total); err != nil {
				return nil, err
			}
			out = append(out, TopItem{ID: id, Count: total})
		} else {
			var id string
			var total uint64
			if err := rows.Scan(&id, &total); err != nil {
				return nil, err
			}
			out = append(out, TopItem{ID: id, Count: int64(total)})
		}
	}
	return out, nil
}
