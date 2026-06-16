package service

import (
	"strings"
)

func Ptr[T any](v T) *T {
	return &v
}

func Capitalize(something string) string {
	something = strings.ToLower(something)

	runes := []rune(something)

	if runes[0] >= 'a' && runes[0] <= 'z' {
		runes[0] = runes[0] - 32
	}

	return string(runes)
}
