package service

import (
	"context"
	"sort"

	"github.com/zahartd/social-network/src/services/stats-service/internal/domain/models"
	"github.com/zahartd/social-network/src/services/stats-service/internal/domain/repository"
)

type Service struct{ repo repository.Stats }

func New(r repository.Stats) *Service { return &Service{repo: r} }

func (s *Service) GetPostStats(ctx context.Context, id string) (int64, int64, int64, error) {
	return s.repo.PostStats(ctx, id)
}

func (s *Service) GetDynamics(ctx context.Context, id, metric string) ([]models.DayCount, error) {
	return s.repo.Dynamics(ctx, id, metric)
}

func (s *Service) GetTopPosts(ctx context.Context, metric string) ([]models.TopItem, error) {
	return s.repo.TopPosts(ctx, metric, 10)
}

func (s *Service) GetTopUsers(ctx context.Context, metric string) ([]models.TopItem, error) {
	items, err := s.repo.TopUsers(ctx, metric, 10)
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].ID < items[j].ID
		}
		return items[i].Count > items[j].Count
	})
	return items, nil
}
