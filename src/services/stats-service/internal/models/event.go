package models

import "time"

type PostLikeEvent struct {
	EventType string    `json:"event_type"`
	UserID    string    `json:"user_id"`
	PostId    string    `json:"post_id"`
	Timestamp time.Time `json:"timestamp"`
}
