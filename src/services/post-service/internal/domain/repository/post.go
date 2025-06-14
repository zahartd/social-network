package repository

import (
	"context"
	"errors"

	"github.com/zahartd/social-network/src/services/post-service/internal/domain/models"
)

var (
	ErrPostNotFound = errors.New("post not found")
	ErrForbidden    = errors.New("forbidden")
)

type Post interface {
	Create(ctx context.Context, p *models.Post) (string, error)
	GetByID(ctx context.Context, id string) (*models.Post, error)
	Update(ctx context.Context, p *models.Post) error
	Delete(ctx context.Context, id, userID string) error

	AuthorID(ctx context.Context, postID string) (string, error)
	UserFeed(ctx context.Context, userID string, page, size int) ([]models.Post, int, error)
	PublicFeed(ctx context.Context, filter *string, page, size int) ([]models.Post, int, error)

	RecordView(ctx context.Context, userID, postID string) error
	RecordLike(ctx context.Context, userID, postID string) error
	RemoveLike(ctx context.Context, userID, postID string) error

	CreateComment(ctx context.Context, c *models.Comment) (string, error)
	CreateReply(ctx context.Context, r *models.Reply) (string, error)
	ListComments(ctx context.Context, postID string, page, size int) ([]models.Comment, int, error)
	ListReplies(ctx context.Context, parentCommentID string) ([]models.Reply, error)
}
