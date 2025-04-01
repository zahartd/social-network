package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/zahartd/social_network/post-service/internal/models"
)

var ErrPostNotFound = errors.New("post not found")
var ErrForbidden = errors.New("forbidden")

type PostRepository interface {
	CreatePost(ctx context.Context, post *models.Post) (string, error)
	GetPostByID(ctx context.Context, postID string) (*models.Post, error)
	UpdatePost(ctx context.Context, post *models.Post) error
	DeletePost(ctx context.Context, postID string, userID string) error
	ListUserPosts(ctx context.Context, userID string, page, pageSize int) ([]models.Post, int, error)
	ListPublicPosts(ctx context.Context, page, pageSize int) ([]models.Post, int, error)
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
		log.Printf("Error creating post: %v", err)
		return "", fmt.Errorf("could not create post: %w", err)
	}
	return postID, nil
}

func (r *postgresPostRepository) GetPostByID(ctx context.Context, postID string) (*models.Post, error) {
	query := `SELECT id, user_id, title, description, created_at, updated_at, is_private, tags FROM posts WHERE id = $1`
	var post models.Post
	err := r.db.GetContext(ctx, &post, query, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		log.Printf("Error getting post by ID %s: %v", postID, err)
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
		log.Printf("Error getting author ID for post %s: %v", postID, err)
		return "", fmt.Errorf("database error fetching author: %w", err)
	}
	return userID, nil
}

func (r *postgresPostRepository) UpdatePost(ctx context.Context, post *models.Post) error {
	query := `UPDATE posts SET title = $1, description = $2, is_private = $3, tags = $4, updated_at = NOW()
              WHERE id = $5`
	result, err := r.db.ExecContext(ctx, query, post.Title, post.Description, post.IsPrivate, post.Tags, post.ID)
	if err != nil {
		log.Printf("Error updating post %s: %v", post.ID, err)
		return fmt.Errorf("could not update post: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected for post update %s: %v", post.ID, err)
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
		log.Printf("Error deleting post %s for user %s: %v", postID, userID, err)
		return fmt.Errorf("could not delete post: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected for post delete %s: %v", postID, err)
		return fmt.Errorf("could not verify post deletion: %w", err)
	}
	if rowsAffected == 0 {
		existsQuery := `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`
		var exists bool
		err := r.db.QueryRowContext(ctx, existsQuery, postID).Scan(&exists)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("Error checking post existence %s: %v", postID, err)
			return fmt.Errorf("could not delete post %s", postID)
		}
		if !exists {
			return ErrPostNotFound
		}
		return ErrForbidden
	}
	return nil
}

func (r *postgresPostRepository) ListUserPosts(ctx context.Context, userID string, page, pageSize int) ([]models.Post, int, error) {
	offset := (page - 1) * pageSize
	query := `SELECT id, user_id, title, description, created_at, updated_at, is_private, tags
              FROM posts
              WHERE user_id = $1
              ORDER BY created_at DESC
              LIMIT $2 OFFSET $3`

	posts := []models.Post{}
	err := r.db.SelectContext(ctx, &posts, query, userID, pageSize, offset)
	if err != nil {
		log.Printf("Error listing user %s posts: %v", userID, err)
		return nil, 0, fmt.Errorf("could not list user posts: %w", err)
	}

	countQuery := `SELECT COUNT(*) FROM posts WHERE user_id = $1`
	var totalCount int
	err = r.db.GetContext(ctx, &totalCount, countQuery, userID)
	if err != nil {
		log.Printf("Error counting user %s posts: %v", userID, err)
		return nil, 0, fmt.Errorf("could not count user posts: %w", err)
	}

	return posts, totalCount, nil
}

func (r *postgresPostRepository) ListPublicPosts(ctx context.Context, page, pageSize int) ([]models.Post, int, error) {
	offset := (page - 1) * pageSize
	query := `SELECT id, user_id, title, description, created_at, updated_at, is_private, tags
              FROM posts
              WHERE is_private = FALSE
              ORDER BY created_at DESC
              LIMIT $1 OFFSET $2`

	posts := []models.Post{}
	err := r.db.SelectContext(ctx, &posts, query, pageSize, offset)
	if err != nil {
		log.Printf("Error listing public posts (page %d, size %d): %v", page, pageSize, err)
		return nil, 0, fmt.Errorf("could not list public posts: %w", err)
	}

	countQuery := `SELECT COUNT(*) FROM posts WHERE is_private = FALSE`
	var totalCount int
	err = r.db.GetContext(ctx, &totalCount, countQuery)
	if err != nil {
		log.Printf("Error counting public posts: %v", err)
		return nil, 0, fmt.Errorf("could not count public posts: %w", err)
	}

	return posts, totalCount, nil
}
