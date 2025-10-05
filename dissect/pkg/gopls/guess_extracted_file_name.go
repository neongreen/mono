package gopls

import (
	"strings"
)

// GuessGoplsExtractedFileName guesses the file name that gopls would use for the extracted function.
//
// TODO(gopls): reuse the guesser from gotools/gopls/internal/golang/extracttofile.go
func GuessGoplsExtractedFileName(funcName string) string {
	return strings.ToLower(funcName) + ".go"
}
