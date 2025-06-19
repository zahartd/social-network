package models

import "time"

type Event struct {
	Metric string
	Time   time.Time
	UserID string
	PostID string
	Count  uint64
}
