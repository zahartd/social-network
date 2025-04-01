package models

import (
	"database/sql/driver"
	"time"

	"github.com/lib/pq"
)

type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	return pq.Array(a).Value()
}

func (a *StringArray) Scan(src interface{}) error {
	return pq.Array(a).Scan(src)
}

type Post struct {
	ID          string      `db:"id"`
	UserID      string      `db:"user_id"`
	Title       string      `db:"title"`
	Description string      `db:"description"`
	CreatedAt   time.Time   `db:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at"`
	IsPrivate   bool        `db:"is_private"`
	Tags        StringArray `db:"tags"`
}
