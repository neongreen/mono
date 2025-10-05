package utils

import (
	"unicode"
)

// isLower checks if a rune is a lowercase letter.
func IsLower(r byte) bool {
	return unicode.IsLower(rune(r))
}
