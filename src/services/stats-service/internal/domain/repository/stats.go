package repository

import (
	"context"

	"github.com/zahartd/social-network/src/services/stats-service/internal/domain/models"
)

type Stats interface {
	Insert(ctx context.Context, ev models.Event) error

	PostStats(ctx context.Context, postID string) (views, likes, comments int64, err error)
	Dynamics(ctx context.Context, postID, metric string) ([]models.DayCount, error)
	TopPosts(ctx context.Context, metric string, limit int) ([]models.TopItem, error)
	TopUsers(ctx context.Context, metric string, limit int) ([]models.TopItem, error)

	Close() error
}
