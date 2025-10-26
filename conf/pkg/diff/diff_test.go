package diff

import (
	"strings"
	"testing"
)

func TestDisplayDiff(t *testing.T) {
	tests := []struct {
		name        string
		before      string
		after       string
		wantChanges bool
	}{
		{
			name:        "no changes",
			before:      "line1\nline2\nline3\n",
			after:       "line1\nline2\nline3\n",
			wantChanges: false,
		},
		{
			name:        "single line change",
			before:      "line1\nline2\nline3\n",
			after:       "line1\nmodified\nline3\n",
			wantChanges: true,
		},
		{
			name:        "addition",
			before:      "line1\nline2\n",
			after:       "line1\nline2\nline3\n",
			wantChanges: true,
		},
		{
			name:        "deletion",
			before:      "line1\nline2\nline3\n",
			after:       "line1\nline3\n",
			wantChanges: true,
		},
		{
			name:        "empty to content",
			before:      "",
			after:       "line1\nline2\n",
			wantChanges: true,
		},
		{
			name:        "content to empty",
			before:      "line1\nline2\n",
			after:       "",
			wantChanges: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasChanges := DisplayDiff(tt.before, tt.after)

			if hasChanges != tt.wantChanges {
				t.Errorf("DisplayDiff() hasChanges = %v, want %v", hasChanges, tt.wantChanges)
			}
		})
	}
}

func TestDisplayUnifiedDiff(t *testing.T) {
	tests := []struct {
		name        string
		before      string
		after       string
		filename    string
		wantChanges bool
	}{
		{
			name:        "no changes",
			before:      "line1\nline2\n",
			after:       "line1\nline2\n",
			filename:    "test.txt",
			wantChanges: false,
		},
		{
			name:        "has changes",
			before:      "line1\nline2\n",
			after:       "line1\nmodified\n",
			filename:    "test.txt",
			wantChanges: true,
		},
		{
			name:        "toml config change",
			before:      "[user]\nname = \"Old\"\nemail = \"old@example.com\"\n",
			after:       "[user]\nname = \"New\"\nemail = \"new@example.com\"\n",
			filename:    "config.toml",
			wantChanges: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasChanges := DisplayUnifiedDiff(tt.before, tt.after, tt.filename)

			if hasChanges != tt.wantChanges {
				t.Errorf("DisplayUnifiedDiff() hasChanges = %v, want %v", hasChanges, tt.wantChanges)
			}
		})
	}
}

func TestDisplayUnifiedDiffReturnsCorrectly(t *testing.T) {
	// Test that it correctly identifies when there are no changes
	before := strings.Repeat("same line\n", 10)
	after := before

	if DisplayUnifiedDiff(before, after, "test.txt") {
		t.Error("DisplayUnifiedDiff should return false when there are no changes")
	}

	// Test that it correctly identifies when there are changes
	before = "line 1\nline 2\nline 3\n"
	after = "line 1\nchanged\nline 3\n"

	if !DisplayUnifiedDiff(before, after, "test.txt") {
		t.Error("DisplayUnifiedDiff should return true when there are changes")
	}
}
