package utils

import (
	"regexp"
	"unicode"
)

func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
func IsNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
func IsStrongPassword(password string) bool {
	var hasLetter, hasNumber, hasSymbol bool

	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsNumber(r):
			hasNumber = true
		case unicode.IsSymbol(r) || unicode.IsPunct(r):
			hasSymbol = true
		}
	}

	return hasLetter && hasNumber && hasSymbol
}
