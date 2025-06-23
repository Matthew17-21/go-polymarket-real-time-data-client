package jsonutils

import "strings"

// IsJsonFormat checks if a string is a valid JSON format
func IsJsonFormat(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}

	// Check if the string is a valid JSON format
	start := s[0]
	end := s[len(s)-1]
	return (start == '{' && end == '}') || (start == '[' && end == ']')
}
