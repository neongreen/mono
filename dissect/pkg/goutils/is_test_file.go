package goutils

import (
	"strings"
)

// isTestFile determines if a given file path corresponds to a Go test file.
func IsTestFile(filePath string) bool {
	return strings.HasSuffix(filePath, "_test.go")
}
