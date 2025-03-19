package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		valid bool
	}{
		{"Valid international", "+1234567890", true},
		{"Valid without plus", "1234567890", true},
		{"Too short", "12345", false},
		{"Too long", "12345678901234567", false},
		{"Invalid characters", "abc1234567", false},
		{"Invalid characters", ";hello world", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		result := ValidatePhone(tt.phone)
		if result != tt.valid {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.valid, result)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"Valid password", "Password123", true},
		{"Too short", "Pwd1", false},
		{"No uppercase", "password123", false},
		{"No lowercase", "PASSWORD123", false},
		{"No digits", "Password", false},
		{"Too long", "Password12345678901234567890", false},
	}

	for _, tt := range tests {
		result := ValidatePassword(tt.password)
		if result != tt.valid {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.valid, result)
		}
	}
}

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name  string
		login string
		valid bool
	}{
		{"Valid login simple", "a123", true},
		{"Valid login with underscore", "john_doe", true},
		{"Valid login numeric in middle", "user123_name", true},
		{"Too short", "a", false},
		{"Too long", "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", false},
		{"Starts with digit", "1john", false},
		{"Starts with underscore", "_john", false},
		{"Contains uppercase", "johnDoe", false},
		{"Contains special char", "john#doe", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		result := ValidateLogin(tt.login)
		if result != tt.valid {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.valid, result)
		}
	}
}

func TestValidateUserID(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		valid  bool
	}{
		{"Valid UUID", uuid.NewString(), true},
		{"Invalid UUID", "not-a-uuid", false},
		{"Empty string", "", false},
		{"Short string", "12345", false},
	}

	for _, tc := range tests {
		result := ValidateUserID(tc.userID)
		if result != tc.valid {
			t.Errorf("%s: expected %v, got %v", tc.name, tc.valid, result)
		}
	}
}
