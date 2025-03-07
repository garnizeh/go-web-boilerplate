package validator

import (
	"net/mail"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

func SanitizeString(input string) string {
	return bluemonday.StrictPolicy().Sanitize(strings.TrimSpace(input))
}

func IsValidEmail(input string) bool {
	if _, err := mail.ParseAddress(input); err != nil {
		return false
	}

	return true
}
