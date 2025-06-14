package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/zahartd/social-network/src/services/post-service/internal/domain/models"
	"github.com/zahartd/social-network/src/services/post-service/internal/domain/repository"
)

type repo struct{ db *sqlx.DB }

func New(db *sqlx.DB) repository.Post { return &repo{db: db} }

/* ------------------- posts ------------------- */

func (r *repo) Create(ctx context.Context, p *models.Post) (string, error) {
	q := `INSERT INTO posts (user_id,title,description,is_private,tags)
	      VALUES ($1,$2,$3,$4,$5) RETURNING id`
	var id string
	if err := r.db.QueryRowContext(ctx, q, p.UserID, p.Title, p.Description, p.IsPrivate, p.Tags).
		Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

func (r *repo) GetByID(ctx context.Context, id string) (*models.Post, error) {
	var p models.Post
	q := `SELECT id,user_id,title,description,created_at,updated_at,is_private,tags
	      FROM posts WHERE id=$1`
	if err := r.db.GetContext(ctx, &p, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrPostNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *repo) Update(ctx context.Context, p *models.Post) error {
	q := `UPDATE posts SET title=$1,description=$2,is_private=$3,tags=$4,updated_at=NOW()
	      WHERE id=$5`
	res, err := r.db.ExecContext(ctx, q, p.Title, p.Description, p.IsPrivate, p.Tags, p.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repository.ErrPostNotFound
	}
	return nil
}

func (r *repo) Delete(ctx context.Context, id, uid string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE id=$1 AND user_id=$2`, id, uid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		var exists bool
		_ = r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM posts WHERE id=$1)`, id)
		if !exists {
			return repository.ErrPostNotFound
		}
		return repository.ErrForbidden
	}
	return nil
}

/* ------------------- helpers & feeds ------------------- */

func (r *repo) AuthorID(ctx context.Context, postID string) (string, error) {
	var uid string
	if err := r.db.GetContext(ctx, &uid, `SELECT user_id FROM posts WHERE id=$1`, postID); err != nil {
		if err == sql.ErrNoRows {
			return "", repository.ErrPostNotFound
		}
		return "", err
	}
	return uid, nil
}

func (r *repo) UserFeed(ctx context.Context, uid string, page, size int) ([]models.Post, int, error) {
	offset := (page - 1) * size
	posts := []models.Post{}
	if err := r.db.SelectContext(ctx, &posts,
		`SELECT id,user_id,title,description,created_at,updated_at,is_private,tags
		   FROM posts WHERE user_id=$1
		   ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		uid, size, offset); err != nil {
		return nil, 0, err
	}
	var total int
	_ = r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM posts WHERE user_id=$1`, uid)
	return posts, total, nil
}

func (r *repo) PublicFeed(ctx context.Context, filter *string, page, size int) ([]models.Post, int, error) {
	offset := (page - 1) * size

	args := []any{}
	where := `is_private = FALSE`
	if filter != nil && *filter != "" {
		where += ` AND user_id = $1`
		args = append(args, *filter)
	}

	query := fmt.Sprintf(
		`SELECT id,user_id,title,description,created_at,updated_at,is_private,tags
		   FROM posts WHERE %s
		   ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, len(args)+1, len(args)+2,
	)
	args = append(args, size, offset)

	var posts []models.Post
	if err := r.db.SelectContext(ctx, &posts, query, args...); err != nil {
		return nil, 0, err
	}

	countQ := `SELECT COUNT(*) FROM posts WHERE ` + where
	var total int
	if err := r.db.GetContext(ctx, &total, countQ, args[:len(args)-2]...); err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

/* ------------------- metrics ------------------- */

func (r *repo) RecordView(ctx context.Context, uid, pid string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO post_views(user_id,post_id) VALUES($1,$2)`, uid, pid)
	return err
}

func (r *repo) RecordLike(ctx context.Context, uid, pid string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO post_likes(user_id,post_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, uid, pid)
	return err
}

func (r *repo) RemoveLike(ctx context.Context, uid, pid string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM post_likes WHERE user_id=$1 AND post_id=$2`, uid, pid)
	return err
}

/* ------------------- comments & replies ------------------- */

func (r *repo) CreateComment(ctx context.Context, c *models.Comment) (string, error) {
	q := `INSERT INTO comments(post_id,user_id,text) VALUES($1,$2,$3) RETURNING id`
	var id string
	if err := r.db.QueryRowContext(ctx, q, c.PostID, c.UserID, c.Text).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

func (r *repo) CreateReply(ctx context.Context, rp *models.Reply) (string, error) {
	q := `INSERT INTO comments(post_id,parent_comment_id,user_id,text)
	      VALUES($1,$2,$3,$4) RETURNING id`
	var id string
	if err := r.db.QueryRowContext(ctx, q, rp.PostID, rp.ParentCommentID, rp.UserID, rp.Text).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

func (r *repo) ListComments(ctx context.Context, postID string, page, size int) ([]models.Comment, int, error) {
	offset := (page - 1) * size
	var cms []models.Comment
	if err := r.db.SelectContext(ctx, &cms,
		`SELECT id,post_id,user_id,text,created_at
		   FROM comments
		  WHERE post_id=$1 AND parent_comment_id IS NULL
		  ORDER BY created_at DESC
		  LIMIT $2 OFFSET $3`,
		postID, size, offset); err != nil {
		return nil, 0, err
	}
	var total int
	_ = r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM comments WHERE post_id=$1 AND parent_comment_id IS NULL`, postID)
	return cms, total, nil
}

func (r *repo) ListReplies(ctx context.Context, parentID string) ([]models.Reply, error) {
	var reps []models.Reply
	if err := r.db.SelectContext(ctx, &reps,
		`SELECT id,post_id,parent_comment_id,user_id,text,created_at
		   FROM comments
		  WHERE parent_comment_id=$1
		  ORDER BY created_at`,
		parentID); err != nil {
		return nil, err
	}
	return reps, nil
}
