package clickhouse_repo

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/zahartd/social-network/src/services/stats-service/internal/domain/models"
	"github.com/zahartd/social-network/src/services/stats-service/internal/domain/repository"
)

type repo struct{ db clickhouse.Conn }

func New(db clickhouse.Conn) repository.Stats { return &repo{db: db} }

func (r *repo) Insert(ctx context.Context, ev models.Event) error {
	const q = `
		INSERT INTO stats.events (event_time,user_id,entity_id,metric,cnt)
		VALUES (?,?,?,?,?)`
	return r.db.Exec(ctx, q, ev.Time, ev.UserID, ev.PostID, ev.Metric, ev.Count)
}

func (r *repo) PostStats(ctx context.Context, id string) (int64, int64, int64, error) {
	var views, likesAdd, likesSub, comments uint64

	if err := r.db.QueryRow(ctx,
		`SELECT sum(cnt) FROM stats.events WHERE entity_id=? AND metric='post-views'`, id).
		Scan(&views); err != nil {
		return 0, 0, 0, err
	}
	if err := r.db.QueryRow(ctx,
		`SELECT sum(cnt) FROM stats.events WHERE entity_id=? AND metric='post-likes'`, id).
		Scan(&likesAdd); err != nil {
		return 0, 0, 0, err
	}
	if err := r.db.QueryRow(ctx,
		`SELECT sum(cnt) FROM stats.events WHERE entity_id=? AND metric='post-unlikes'`, id).
		Scan(&likesSub); err != nil {
		return 0, 0, 0, err
	}
	if err := r.db.QueryRow(ctx,
		`SELECT sum(cnt) FROM stats.events WHERE entity_id=? AND metric='post-comments'`, id).
		Scan(&comments); err != nil {
		return 0, 0, 0, err
	}

	return int64(views), int64(likesAdd) - int64(likesSub), int64(comments), nil
}

func (r *repo) Dynamics(ctx context.Context, id, metric string) ([]models.DayCount, error) {
	rows, err := r.db.Query(ctx, `
		SELECT toDate(event_time) d, sum(cnt) c
		FROM stats.events
		WHERE entity_id=? AND metric=?
		GROUP BY d ORDER BY d`,
		id, metric)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.DayCount
	for rows.Next() {
		var d time.Time
		var c uint64
		if err := rows.Scan(&d, &c); err != nil {
			return nil, err
		}
		out = append(out, models.DayCount{
			Date:  d.Format("2006-01-02"),
			Count: c,
		})
	}
	return out, nil
}

func (r *repo) TopPosts(ctx context.Context, metric string, limit int) ([]models.TopItem, error) {
	var sql string
	args := make([]any, 0, 2)
	signed := false

	if metric == "post-likes" {
		sql = `
			SELECT entity_id,
			       sumIf(cnt,metric='post-likes') -
			       sumIf(cnt,metric='post-unlikes') AS total
			FROM stats.events
			GROUP BY entity_id
			ORDER BY total DESC
			LIMIT ?`
		args = append(args, limit)
		signed = true
	} else {
		sql = `
			SELECT entity_id, sum(cnt) AS total
			FROM stats.events
			WHERE metric = ?
			GROUP BY entity_id
			ORDER BY total DESC
			LIMIT ?`
		args = append(args, metric, limit)
	}

	return r.selectTop(ctx, sql, signed, args...)
}

func (r *repo) TopUsers(ctx context.Context, metric string, limit int) ([]models.TopItem, error) {
	var sql string
	args := make([]any, 0, 2)
	signed := false

	if metric == "post-likes" {
		sql = `
			SELECT user_id,
			       sumIf(cnt,metric='post-likes') -
			       sumIf(cnt,metric='post-unlikes') AS total
			FROM stats.events
			GROUP BY user_id
			ORDER BY total DESC
			LIMIT ?`
		args = append(args, limit)
		signed = true
	} else {
		sql = `
			SELECT user_id, sum(cnt) AS total
			FROM stats.events
			WHERE metric = ?
			GROUP BY user_id
			ORDER BY total DESC
			LIMIT ?`
		args = append(args, metric, limit)
	}

	return r.selectTop(ctx, sql, signed, args...)
}

func (r *repo) selectTop(ctx context.Context, sql string, signed bool, args ...any) ([]models.TopItem, error) {
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.TopItem
	for rows.Next() {
		var id string
		if signed {
			var cnt int64
			if err := rows.Scan(&id, &cnt); err != nil {
				return nil, err
			}
			items = append(items, models.TopItem{ID: id, Count: cnt})
		} else {
			var cnt uint64
			if err := rows.Scan(&id, &cnt); err != nil {
				return nil, err
			}
			items = append(items, models.TopItem{ID: id, Count: int64(cnt)})
		}
	}
	return items, nil
}

func (r *repo) Close() error { return r.db.Close() }
