package internal

import "strings"

func IsPrivate(s string) bool {
	if s == "" {
		return true
	}
	return strings.ToLower(string(s[0])) == string(s[0])
}
