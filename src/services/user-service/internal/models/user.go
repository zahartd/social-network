package models

import "time"

type User struct {
	ID           string    `json:"id"` // uuid
	Login        string    `json:"login"`
	Firstname    string    `json:"firstname"`
	Surname      string    `json:"surname"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone,omitempty"`
	Bio          string    `json:"bio,omitempty"`
	PasswordHash string    `json:"-"`
	Salt         string    `json:"salt,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
