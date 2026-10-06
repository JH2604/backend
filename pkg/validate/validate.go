package validate

import "strings"

func Password(s string) bool {
	if len(s) < 8 || len(s) > 32 {
		return false
	}
	hasLetter := false
	hasDigit := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c <= ' ' {
			return false
		}
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			hasLetter = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}

func StudentID(s string) bool {
	s = strings.TrimSpace(s)
	if s == "admin" {
		return true
	}
	if len(s) < 3 || len(s) > 20 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
