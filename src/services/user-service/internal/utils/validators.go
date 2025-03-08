package utils

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var phoneRegex = regexp.MustCompile(`^\+?[0-9]{7,15}$`)

func PhoneValidator(fl validator.FieldLevel) bool {
	phone, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	return phoneRegex.MatchString(phone)
}

func PasswordValidator(fl validator.FieldLevel) bool {
	password, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	if len(password) < 8 || len(password) > 24 {
		return false
	}

	lower := regexp.MustCompile(`[a-z]`)
	upper := regexp.MustCompile(`[A-Z]`)
	digit := regexp.MustCompile(`[0-9]`)

	return lower.MatchString(password) && upper.MatchString(password) && digit.MatchString(password)
}
