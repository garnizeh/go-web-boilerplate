package validator

import (
	"net/mail"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

func SanitizeString(str string) string {
	str = strings.TrimSpace(str)
	return bluemonday.StrictPolicy().Sanitize(str)
}

func IsValidEmail(email string) bool {
	if _, err := mail.ParseAddress(email); err != nil {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	return strings.Contains(parts[1], ".")
}
