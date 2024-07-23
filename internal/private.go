package internal

import "strings"

func IsPrivate(s string) bool {
	if s == "" {
		return false
	}
	return strings.ToLower(string(s[0])) == string(s[0])
}
