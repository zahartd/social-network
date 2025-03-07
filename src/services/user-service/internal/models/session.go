package models

import "time"

type Session struct {
	ID        string    `json:"id"` // uuid
	UserID    string    `json:"userId"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	IPAddress string    `json:"ipAddress"`
}
