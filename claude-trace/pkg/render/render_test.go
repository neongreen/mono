package render

import (
	"strings"
	"testing"
)

func TestFormatWriteToolArguments(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected string
	}{
		{
			name: "normal write tool arguments",
			input: map[string]any{
				"file_path": "/path/to/file.txt",
				"content":   "Hello, world!",
			},
			expected: "**File:** `/path/to/file.txt`\n\n**Content:**\n```\nHello, world!\n```\n\n",
		},
		{
			name: "write tool with unexpected fields",
			input: map[string]any{
				"file_path":     "/path/to/file.txt",
				"content":       "Hello, world!",
				"extra_field":   "unexpected",
				"another_field": 123,
			},
			expected: "**File:** `/path/to/file.txt`\n\n**Content:**\n```\nHello, world!\n```\n\n⚠️ **Unexpected fields:** `extra_field`, `another_field`\n\n",
		},
		{
			name: "write tool with only file_path",
			input: map[string]any{
				"file_path": "/path/to/file.txt",
			},
			expected: "**File:** `/path/to/file.txt`\n\n",
		},
		{
			name: "write tool with only content",
			input: map[string]any{
				"content": "Hello, world!",
			},
			expected: "**Content:**\n```\nHello, world!\n```\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatWriteToolArguments(tt.input)
			if result != tt.expected {
				t.Errorf("formatWriteToolArguments() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatToolArguments(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		input    map[string]any
		contains string
	}{
		{
			name:     "write tool (lowercase)",
			toolName: "write",
			input: map[string]any{
				"file_path": "/test.txt",
				"content":   "test",
			},
			contains: "**File:**",
		},
		{
			name:     "write tool (uppercase)",
			toolName: "Write",
			input: map[string]any{
				"file_path": "/test.txt",
				"content":   "test",
			},
			contains: "**File:**",
		},
		{
			name:     "other tool",
			toolName: "read_file",
			input: map[string]any{
				"path": "/test.txt",
			},
			contains: "```json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatToolArguments(tt.toolName, tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("formatToolArguments() = %q, should contain %q", result, tt.contains)
			}
		})
	}
}
