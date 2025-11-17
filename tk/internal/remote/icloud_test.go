package remote

import (
	"testing"
)

func TestIsICloudPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "standard iCloud Drive path",
			path:     "/Users/test/Library/Mobile Documents/com~apple~CloudDocs/tk-events",
			expected: true,
		},
		{
			name:     "app-specific iCloud path",
			path:     "/Users/test/Library/Mobile Documents/iCloud~com~example~app/data",
			expected: true,
		},
		{
			name:     "regular home directory path",
			path:     "/Users/test/Documents/tk-events",
			expected: false,
		},
		{
			name:     "regular tmp path",
			path:     "/tmp/tk-events",
			expected: false,
		},
		{
			name:     "root path",
			path:     "/",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isICloudPath(tt.path)
			if result != tt.expected {
				t.Errorf("isICloudPath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}
