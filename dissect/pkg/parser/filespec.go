package parser

import (
	"fmt"
	"strings"
)

// FileSpec represents a parsed file specification in the format "file.go:identifier"
type FileSpec struct {
	// FilePath is the path to the file (required)
	FilePath string
	// Identifier is the symbol name (optional, empty if not specified)
	Identifier string
}

// ParseFileSpec parses a file specification string in the format:
//   - "file.go" -> FileSpec{FilePath: "file.go", Identifier: ""}
//   - "file.go:Foo" -> FileSpec{FilePath: "file.go", Identifier: "Foo"}
//   - "path/to/file.go:MyFunc" -> FileSpec{FilePath: "path/to/file.go", Identifier: "MyFunc"}
//
// Returns an error if the format is invalid (e.g., empty string, only colon, etc.)
func ParseFileSpec(spec string) (FileSpec, error) {
	if spec == "" {
		return FileSpec{}, fmt.Errorf("file specification cannot be empty")
	}

	// Check for colon
	colonIndex := strings.Index(spec, ":")
	if colonIndex == -1 {
		// No colon, just a file path
		return FileSpec{FilePath: spec, Identifier: ""}, nil
	}

	// Split by colon
	filePath := spec[:colonIndex]
	identifier := spec[colonIndex+1:]

	// Validate
	if filePath == "" {
		return FileSpec{}, fmt.Errorf("file path cannot be empty")
	}

	if identifier == "" {
		return FileSpec{}, fmt.Errorf("identifier after ':' cannot be empty (use format 'file.go' or 'file.go:identifier')")
	}

	return FileSpec{FilePath: filePath, Identifier: identifier}, nil
}

// String returns the string representation of the FileSpec
func (fs FileSpec) String() string {
	if fs.Identifier == "" {
		return fs.FilePath
	}
	return fmt.Sprintf("%s:%s", fs.FilePath, fs.Identifier)
}

// HasIdentifier returns true if an identifier is specified
func (fs FileSpec) HasIdentifier() bool {
	return fs.Identifier != ""
}
