package models

import (
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID          string         `db:"id"`
	UserID      string         `db:"user_id"`
	Title       string         `db:"title"`
	Description string         `db:"description"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	IsPrivate   bool           `db:"is_private"`
	Tags        pq.StringArray `db:"tags"`
}
