package service

import (
	"strings"

	"github.com/google/uuid"
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

func CheckUUID(IDs ...uuid.UUID) bool {
	length := len(IDs)

	for i := 0; i < length; i++ {
		if IDs[i] == uuid.Nil {
			return false
		}
	}

	return true
}
