package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/zahartd/social-network/src/services/post-service/internal/models"
)

var ErrPostNotFound = errors.New("post not found")
var ErrForbidden = errors.New("forbidden")

type PostRepository interface {
	CreatePost(ctx context.Context, post *models.Post) (string, error)
	GetPostByID(ctx context.Context, postID string) (*models.Post, error)
	UpdatePost(ctx context.Context, post *models.Post) error
	DeletePost(ctx context.Context, postID string, userID string) error
	GetUserPosts(ctx context.Context, userID string, page, pageSize int) ([]models.Post, int, error)
	GetPublicPosts(ctx context.Context, filterUserID *string, page, pageSize int) ([]models.Post, int, error)
	GetPostAuthorID(ctx context.Context, postID string) (string, error)
}

type postgresPostRepository struct {
	db *sqlx.DB
}

func NewPostgresPostRepository(db *sqlx.DB) PostRepository {
	return &postgresPostRepository{db: db}
}

func (r *postgresPostRepository) CreatePost(ctx context.Context, post *models.Post) (string, error) {
	query := `INSERT INTO posts (user_id, title, description, is_private, tags)
              VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var postID string
	err := r.db.QueryRowContext(ctx, query, post.UserID, post.Title, post.Description, post.IsPrivate, post.Tags).Scan(&postID)
	if err != nil {
		return "", fmt.Errorf("could not create post: %w", err)
	}
	return postID, nil
}

func (r *postgresPostRepository) GetPostByID(ctx context.Context, postID string) (*models.Post, error) {
	query := `SELECT id, user_id, title, description, created_at, updated_at, is_private, tags FROM posts WHERE id = $1`
	var post models.Post
	err := r.db.Get(&post, query, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("could not get post: %w", err)
	}
	return &post, nil
}

func (r *postgresPostRepository) GetPostAuthorID(ctx context.Context, postID string) (string, error) {
	query := `SELECT user_id FROM posts WHERE id = $1`
	var userID string
	err := r.db.GetContext(ctx, &userID, query, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrPostNotFound
		}
		return "", fmt.Errorf("database error fetching author: %w", err)
	}
	return userID, nil
}

func (r *postgresPostRepository) UpdatePost(ctx context.Context, post *models.Post) error {
	query := `UPDATE posts SET title = $1, description = $2, is_private = $3, tags = $4, updated_at = NOW()
              WHERE id = $5`
	result, err := r.db.ExecContext(ctx, query, post.Title, post.Description, post.IsPrivate, post.Tags, post.ID)
	if err != nil {
		return fmt.Errorf("could not update post: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not verify post update: %w", err)
	}
	if rowsAffected == 0 {
		return ErrPostNotFound
	}
	return nil
}

func (r *postgresPostRepository) DeletePost(ctx context.Context, postID string, userID string) error {
	query := `DELETE FROM posts WHERE id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, postID, userID)
	if err != nil {
		return fmt.Errorf("could not delete post: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not verify post deletion: %w", err)
	}
	if rowsAffected == 0 {
		existsQuery := `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`
		var exists bool
		err := r.db.QueryRowContext(ctx, existsQuery, postID).Scan(&exists)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("could not delete post %s", postID)
		}
		if !exists {
			return ErrPostNotFound
		}
		return ErrForbidden
	}
	return nil
}

func (r *postgresPostRepository) GetUserPosts(ctx context.Context, userID string, page, pageSize int) ([]models.Post, int, error) {
	offset := (page - 1) * pageSize
	query := `SELECT id, user_id, title, description, created_at, updated_at, is_private, tags
              FROM posts
              WHERE user_id = $1
              ORDER BY created_at DESC
              LIMIT $2 OFFSET $3`

	posts := []models.Post{}
	err := r.db.SelectContext(ctx, &posts, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("could not list user posts: %w", err)
	}

	countQuery := `SELECT COUNT(*) FROM posts WHERE user_id = $1`
	var totalCount int
	err = r.db.GetContext(ctx, &totalCount, countQuery, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("could not count user posts: %w", err)
	}

	return posts, totalCount, nil
}

func (r *postgresPostRepository) GetPublicPosts(ctx context.Context, filterUserID *string, page, pageSize int) ([]models.Post, int, error) {
	offset := (page - 1) * pageSize
	args := []any{}
	countArgs := []any{}

	queryBase := `SELECT id, user_id, title, description, created_at, updated_at, is_private, tags FROM posts`
	countQueryBase := `SELECT COUNT(*) FROM posts`
	whereClause := ` WHERE is_private = FALSE`

	paramIndex := 1
	if filterUserID != nil && *filterUserID != "" {
		whereClause += fmt.Sprintf(" AND user_id = $%d", paramIndex)
		args = append(args, *filterUserID)
		countArgs = append(countArgs, *filterUserID)
		paramIndex++
	}

	query := queryBase + whereClause + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", paramIndex, paramIndex+1)
	args = append(args, pageSize, offset)
	countQuery := countQueryBase + whereClause

	posts := []models.Post{}
	err := r.db.SelectContext(ctx, &posts, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("database query error: %w", err)
	}

	var totalCount int
	err = r.db.GetContext(ctx, &totalCount, countQuery, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("database count query error: %w", err)
	}
	return posts, totalCount, nil
}
