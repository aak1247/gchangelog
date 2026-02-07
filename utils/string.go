package utils

import (
	"strings"
)

func IsMultiline(s string) bool {
	return strings.Contains(s, "\n")
}
