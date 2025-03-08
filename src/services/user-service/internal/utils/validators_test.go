package utils

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

type TestStruct struct {
	Phone string `validate:"phone"`
}

func TestPhoneValidator(t *testing.T) {
	v := validator.New()
	if err := v.RegisterValidation("phone", PhoneValidator); err != nil {
		t.Fatalf("Failed to register phone validator: %v", err)
	}

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
		ts := TestStruct{Phone: tt.phone}
		err := v.Struct(ts)
		if (err == nil) != tt.valid {
			t.Errorf("%s: expected valid=%v, got error: %v", tt.name, tt.valid, err)
		} else if !tt.valid && err == nil {
			t.Errorf("Expected error for %q, but got valid", tt.name)
		}
	}
}

type TestPassword struct {
	Password string `validate:"password"`
}

func TestPasswordValidator(t *testing.T) {
	v := validator.New()
	if err := v.RegisterValidation("password", PasswordValidator); err != nil {
		t.Fatalf("Failed to register password validator: %v", err)
	}

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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := TestPassword{Password: tc.password}
			err := v.Struct(data)
			if tc.valid && err != nil {
				t.Errorf("Expected valid for %q, got error: %v", tc.password, err)
			} else if !tc.valid && err == nil {
				t.Errorf("Expected error for %q, but got valid", tc.password)
			}
		})
	}
}
