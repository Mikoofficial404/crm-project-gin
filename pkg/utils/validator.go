package utils

import (
	"regexp"
	"strings"
)

var phoneRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var nonDigitRegex = regexp.MustCompile(`\D`)

func IsValidEmail(email string) bool {
	return phoneRegex.MatchString(email)
}

func NormalizePhone(phone string) string {

	digits := nonDigitRegex.ReplaceAllString(phone, "")

	if digits == "" {
		return phone
	}

	if strings.HasPrefix(digits, "62") {
		return digits
	}

	if strings.HasPrefix(digits, "0") {
		return "62" + digits[1:]
	}

	if strings.HasPrefix(digits, "8") {
		return "62" + digits
	}

	return digits
}
