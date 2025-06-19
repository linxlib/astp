package internal

import (
	"github.com/linxlib/astp/constants"
	"strings"
)

func IsPrivate(s string) bool {
	if s == "" || s == constants.EmptyName {
		return true
	}
	return strings.ToLower(string(s[0])) == string(s[0])
}
