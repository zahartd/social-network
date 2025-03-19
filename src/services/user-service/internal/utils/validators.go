package utils

import (
	"regexp"

	"github.com/google/uuid"
)

func ValidatePhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{7,15}$`)
	return phoneRegex.MatchString(phone)
}

func ValidatePassword(password string) bool {
	if len(password) < 8 || len(password) > 24 {
		return false
	}
	lowerRegex := regexp.MustCompile(`[a-z]`)
	upperRegex := regexp.MustCompile(`[A-Z]`)
	digitRegex := regexp.MustCompile(`[0-9]`)
	return lowerRegex.MatchString(password) && upperRegex.MatchString(password) && digitRegex.MatchString(password)
}

func ValidateLogin(login string) bool {
	if len(login) < 2 || len(login) > 40 {
		return false
	}
	loginRegex := regexp.MustCompile(`^[a-z][a-z0-9_]+$`)
	return loginRegex.MatchString(login)
}

func ValidateUserID(userID string) bool {
	_, err := uuid.Parse(userID)
	return err == nil
}
