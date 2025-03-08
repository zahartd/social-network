package utils

import (
	"fmt"
	"strings"
)

func ParseIdentifier(identifier string) (string, bool, error) {
	if strings.Contains(identifier, "-") {
		if !ValidateUserID(identifier) {
			return "", false, fmt.Errorf("invalid UUID format")
		}
		return identifier, true, nil
	} else {
		if !ValidateLogin(identifier) {
			return "", false, fmt.Errorf("invalid login")
		}
		return identifier, false, nil
	}
}
