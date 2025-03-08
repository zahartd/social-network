package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestParseIdentifierValidUUID(t *testing.T) {
	validUUID := uuid.NewString()
	id, isUUID, err := ParseIdentifier(validUUID)
	if err != nil {
		t.Errorf("Expected no error for valid UUID, got: %v", err)
	}
	if !isUUID {
		t.Errorf("Expected isUUID=true for valid UUID, got false")
	}
	if id != validUUID {
		t.Errorf("Expected returned id %q, got %q", validUUID, id)
	}
}

func TestParseIdentifierInvalidUUID(t *testing.T) {
	invalidUUID := "12345-67890"
	_, isUUID, err := ParseIdentifier(invalidUUID)
	if err == nil {
		t.Errorf("Expected error for invalid UUID, got nil")
	}
	if isUUID {
		t.Errorf("Expected isUUID=false for invalid UUID, got true")
	}
	expectedErr := "invalid UUID format"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestParseIdentifierValidLogin(t *testing.T) {
	validLogin := "john_doe"
	id, isUUID, err := ParseIdentifier(validLogin)
	if err != nil {
		t.Errorf("Expected no error for valid login, got: %v", err)
	}
	if isUUID {
		t.Errorf("Expected isUUID=false for valid login, got true")
	}
	if id != validLogin {
		t.Errorf("Expected returned id %q, got %q", validLogin, id)
	}
}

func TestParseIdentifier_InvalidLogin(t *testing.T) {
	invalidLogin := "JohnDoe"
	_, isUUID, err := ParseIdentifier(invalidLogin)
	if err == nil {
		t.Errorf("Expected error for invalid login, got nil")
	}
	if isUUID {
		t.Errorf("Expected isUUID=false for invalid login, got true")
	}
	expectedErr := "invalid login"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
	}
}
