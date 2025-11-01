package parser

import (
	"testing"
)

func TestParseFileSpec(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      FileSpec
		expectError   bool
		errorContains string
	}{
		{
			name:     "simple file",
			input:    "file.go",
			expected: FileSpec{FilePath: "file.go", Identifier: ""},
		},
		{
			name:     "file with identifier",
			input:    "file.go:MyFunc",
			expected: FileSpec{FilePath: "file.go", Identifier: "MyFunc"},
		},
		{
			name:     "path with identifier",
			input:    "path/to/file.go:MyType",
			expected: FileSpec{FilePath: "path/to/file.go", Identifier: "MyType"},
		},
		{
			name:     "deep path with identifier",
			input:    "pkg/internal/utils/helper.go:HelperFunc",
			expected: FileSpec{FilePath: "pkg/internal/utils/helper.go", Identifier: "HelperFunc"},
		},
		{
			name:     "relative path",
			input:    "./file.go:Func",
			expected: FileSpec{FilePath: "./file.go", Identifier: "Func"},
		},
		{
			name:     "parent directory path",
			input:    "../other/file.go:Func",
			expected: FileSpec{FilePath: "../other/file.go", Identifier: "Func"},
		},
		{
			name:     "absolute path",
			input:    "/usr/local/src/file.go:Main",
			expected: FileSpec{FilePath: "/usr/local/src/file.go", Identifier: "Main"},
		},
		{
			name:     "glob pattern file",
			input:    "*.go",
			expected: FileSpec{FilePath: "*.go", Identifier: ""},
		},
		{
			name:     "glob pattern with identifier",
			input:    "*.go:Helper",
			expected: FileSpec{FilePath: "*.go", Identifier: "Helper"},
		},
		{
			name:     "doublestar glob pattern",
			input:    "pkg/**/*.go:Test*",
			expected: FileSpec{FilePath: "pkg/**/*.go", Identifier: "Test*"},
		},
		{
			name:     "identifier with underscore",
			input:    "file.go:my_func",
			expected: FileSpec{FilePath: "file.go", Identifier: "my_func"},
		},
		{
			name:     "identifier with numbers",
			input:    "file.go:Func123",
			expected: FileSpec{FilePath: "file.go", Identifier: "Func123"},
		},
		{
			name:          "empty string",
			input:         "",
			expectError:   true,
			errorContains: "cannot be empty",
		},
		{
			name:          "only colon",
			input:         ":",
			expectError:   true,
			errorContains: "file path cannot be empty",
		},
		{
			name:          "empty file path",
			input:         ":Identifier",
			expectError:   true,
			errorContains: "file path cannot be empty",
		},
		{
			name:          "empty identifier after colon",
			input:         "file.go:",
			expectError:   true,
			errorContains: "identifier after ':' cannot be empty",
		},
		{
			name:          "colon at start",
			input:         ":func",
			expectError:   true,
			errorContains: "file path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFileSpec(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.FilePath != tt.expected.FilePath {
				t.Errorf("FilePath mismatch: got %q, want %q", result.FilePath, tt.expected.FilePath)
			}

			if result.Identifier != tt.expected.Identifier {
				t.Errorf("Identifier mismatch: got %q, want %q", result.Identifier, tt.expected.Identifier)
			}
		})
	}
}

func TestFileSpec_String(t *testing.T) {
	tests := []struct {
		name     string
		spec     FileSpec
		expected string
	}{
		{
			name:     "file only",
			spec:     FileSpec{FilePath: "file.go", Identifier: ""},
			expected: "file.go",
		},
		{
			name:     "file with identifier",
			spec:     FileSpec{FilePath: "file.go", Identifier: "Func"},
			expected: "file.go:Func",
		},
		{
			name:     "path with identifier",
			spec:     FileSpec{FilePath: "pkg/file.go", Identifier: "MyType"},
			expected: "pkg/file.go:MyType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFileSpec_HasIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		spec     FileSpec
		expected bool
	}{
		{
			name:     "no identifier",
			spec:     FileSpec{FilePath: "file.go", Identifier: ""},
			expected: false,
		},
		{
			name:     "has identifier",
			spec:     FileSpec{FilePath: "file.go", Identifier: "Func"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.HasIdentifier()
			if result != tt.expected {
				t.Errorf("HasIdentifier() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseFileSpec_MultipleColons(t *testing.T) {
	// Test that only the first colon is used as the separator
	// This handles cases like Windows paths or URLs
	result, err := ParseFileSpec("file.go:Func:Extra")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.FilePath != "file.go" {
		t.Errorf("FilePath = %q, want %q", result.FilePath, "file.go")
	}

	// The identifier should include everything after the first colon
	if result.Identifier != "Func:Extra" {
		t.Errorf("Identifier = %q, want %q", result.Identifier, "Func:Extra")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && findSubstr(s, substr)))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
