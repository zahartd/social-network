package models

import "time"

type Reply struct {
	ID              string    `db:"id"`
	PostID          string    `db:"post_id"`
	UserID          string    `db:"user_id"`
	ParentCommentID string    `db:"parent_comment_id"`
	Text            string    `db:"text"`
	CreatedAt       time.Time `db:"created_at"`
}
