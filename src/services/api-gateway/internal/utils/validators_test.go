package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestValidateUserID(t *testing.T) {
	tests := []struct {
		name        string
		userIDValue any
		valid       error
	}{
		{"Valid UUID", uuid.NewString(), nil},
		{"Invalid UUID", "not-a-uuid", ErrInvalidUserID},
		{"Empty string", "", ErrInvalidUserID},
		{"Short string", "12345", ErrInvalidUserID},
		{"Non string", 1, ErrInvalidUserID},
	}

	for _, tc := range tests {
		result := ValidateUserID(tc.userIDValue)
		if result != tc.valid {
			t.Errorf("%s: expected %v, got %v", tc.name, tc.valid, result)
		}
	}
}

func TestValidatePage(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{"valid page", "5", 5, false},
		{"invalid non-numeric", "abc", 0, true},
		{"zero", "0", 0, true},
		{"negative", "-3", 0, true},
		{"minimum valid", "1", 1, false},
		{"large number", "100000", 100000, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidatePage(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidatePage(%q) got error %v, wantErr %t", tc.input, err, tc.wantErr)
			}
			if got != tc.expected {
				t.Errorf("ValidatePage(%q) = %d, want %d", tc.input, got, tc.expected)
			}
		})
	}
}

func TestValidatePageSize(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{"valid page size", "10", 10, false},
		{"invalid non-numeric", "xyz", 0, true},
		{"zero", "0", 0, true},
		{"negative", "-5", 0, true},
		{"minimum valid", "1", 1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidatePageSize(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidatePageSize(%q) got error %v, wantErr %t", tc.input, err, tc.wantErr)
			}
			if got != tc.expected {
				t.Errorf("ValidatePageSize(%q) = %d, want %d", tc.input, got, tc.expected)
			}
		})
	}
}

func TestValidatePostID(t *testing.T) {
	testCases := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{"valid UUID", "123e4567-e89b-12d3-a456-426614174000", false},
		{"invalid format (short)", "123", true},
		{"invalid characters", "123z-...", true},
		{"empty string", "", true},
		{"non-string type (int)", 42, true},
		{"non-string type (nil)", nil, true},
		{"non-string type (bool)", true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePostID(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidatePostID(%v) error = %v, wantErr %t", tc.input, err, tc.wantErr)
			}
			if tc.wantErr && err != ErrInvalidPostID {
				t.Errorf("Expected error %v, got %v", ErrInvalidPostID, err)
			}
		})
	}
}
