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
	RecordView(ctx context.Context, userID, postID string) error
	RecordLike(ctx context.Context, userID, postID string) error
	RemoveLike(ctx context.Context, userID, postID string) error
	CreateComment(ctx context.Context, cm *models.Comment) (string, error)
	ListComments(ctx context.Context, postID string, page, pageSize int) ([]models.Comment, int, error)
	ListReplies(ctx context.Context, parentCommentID string) ([]models.Comment, error)
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

func (r *postgresPostRepository) RecordView(ctx context.Context, userID, postID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO post_views (user_id, post_id) VALUES ($1,$2)`, userID, postID)
	return err
}

func (r *postgresPostRepository) RecordLike(ctx context.Context, userID, postID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO post_likes (user_id, post_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, userID, postID)
	return err
}

func (r *postgresPostRepository) RemoveLike(ctx context.Context, userID, postID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM post_likes WHERE user_id=$1 AND post_id=$2`, userID, postID)
	return err
}

func (r *postgresPostRepository) CreateComment(ctx context.Context, cm *models.Comment) (string, error) {
	query := `INSERT INTO comments (post_id, parent_comment_id, user_id, text) VALUES ($1,$2,$3,$4) RETURNING id`
	var id string
	err := r.db.QueryRowContext(ctx, query, cm.PostID, cm.ParentCommentID, cm.UserID, cm.Text).Scan(&id)
	return id, err
}

func (r *postgresPostRepository) ListComments(ctx context.Context, postID string, page, pageSize int) ([]models.Comment, int, error) {
	offset := (page - 1) * pageSize
	comments := []models.Comment{}
	err := r.db.SelectContext(
		ctx,
		&comments,
		`SELECT id, post_id, parent_comment_id, user_id, text, created_at
		   FROM comments
		  WHERE post_id = $1
		    AND parent_comment_id IS NULL
		  ORDER BY created_at DESC
		  LIMIT $2 OFFSET $3`,
		postID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("could not list comments: %w", err)
	}

	var total int
	err = r.db.GetContext(ctx,
		&total,
		`SELECT COUNT(*)
		   FROM comments
		  WHERE post_id = $1
		    AND parent_comment_id IS NULL`,
		postID,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("could not count comments: %w", err)
	}

	return comments, total, nil
}

func (r *postgresPostRepository) ListReplies(ctx context.Context, parentID string) ([]models.Comment, error) {
	replies := []models.Comment{}
	err := r.db.SelectContext(
		ctx,
		&replies,
		`SELECT id, post_id, parent_comment_id, user_id, text, created_at
		   FROM comments
		  WHERE parent_comment_id = $1
		  ORDER BY created_at`,
		parentID,
	)
	if err != nil {
		return nil, fmt.Errorf("could not list replies: %w", err)
	}
	return replies, nil
}
