package inmemory

import (
	"context"
	"sort"
	"sync"

	"github.com/zahartd/social-network/src/services/stats-service/internal/domain/models"
	"github.com/zahartd/social-network/src/services/stats-service/internal/domain/repository"
)

type repo struct {
	mu     sync.RWMutex
	events []models.Event
}

func New() repository.Stats { return &repo{} }

func (r *repo) Insert(_ context.Context, ev models.Event) error {
	r.mu.Lock()
	r.events = append(r.events, ev)
	r.mu.Unlock()
	return nil
}

func aggregate(events []models.Event, keyFn func(models.Event) string) map[string]int64 {
	out := map[string]int64{}
	for _, e := range events {
		k := keyFn(e)
		switch e.Metric {
		case "post-likes":
			out[k] += int64(e.Count)
		case "post-unlikes":
			out[k] -= int64(e.Count)
		default: // views & comments
			out[k] += int64(e.Count)
		}
	}
	return out
}

func topN(m map[string]int64, n int) []models.TopItem {
	list := make([]models.TopItem, 0, len(m))
	for id, c := range m {
		list = append(list, models.TopItem{ID: id, Count: c})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Count == list[j].Count {
			return list[i].ID < list[j].ID
		}
		return list[i].Count > list[j].Count
	})
	if len(list) > n {
		list = list[:n]
	}
	return list
}

func (r *repo) PostStats(_ context.Context, postID string) (int64, int64, int64, error) {
	var views, likes, comments int64

	r.mu.RLock()
	for _, e := range r.events {
		if e.PostID != postID {
			continue
		}
		switch e.Metric {
		case "post-views":
			views += int64(e.Count)
		case "post-likes":
			likes += int64(e.Count)
		case "post-unlikes":
			likes -= int64(e.Count)
		case "post-comments":
			comments += int64(e.Count)
		}
	}
	r.mu.RUnlock()

	return views, likes, comments, nil
}

func (r *repo) Dynamics(_ context.Context, postID, metric string) ([]models.DayCount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	daily := map[string]uint64{}

	for _, e := range r.events {
		if e.PostID != postID {
			continue
		}
		if metric == "post-likes" {
			if e.Metric != "post-likes" && e.Metric != "post-unlikes" {
				continue
			}
		} else if e.Metric != metric {
			continue
		}
		day := e.Time.Format("2006-01-02")
		daily[day] += e.Count
	}

	out := make([]models.DayCount, 0, len(daily))
	for d, c := range daily {
		out = append(out, models.DayCount{Date: d, Count: c})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date < out[j].Date })
	return out, nil
}

func (r *repo) TopPosts(_ context.Context, metric string, limit int) ([]models.TopItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []models.Event
	for _, e := range r.events {
		if metric == "post-likes" {
			if e.Metric != "post-likes" && e.Metric != "post-unlikes" {
				continue
			}
		} else if e.Metric != metric {
			continue
		}
		filtered = append(filtered, e)
	}

	agg := aggregate(filtered, func(e models.Event) string { return e.PostID })
	return topN(agg, limit), nil
}

func (r *repo) TopUsers(_ context.Context, metric string, limit int) ([]models.TopItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []models.Event
	for _, e := range r.events {
		if metric == "post-likes" {
			if e.Metric != "post-likes" && e.Metric != "post-unlikes" {
				continue
			}
		} else if e.Metric != metric {
			continue
		}
		filtered = append(filtered, e)
	}

	agg := aggregate(filtered, func(e models.Event) string { return e.UserID })
	return topN(agg, limit), nil
}

func (r *repo) Close() error { return nil }
