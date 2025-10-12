package main

import (
	"testing"
)

func TestExtractProjectFromTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want string
	}{
		{
			name: "PR release with double dash",
			tag:  "dissect--pr-123.1",
			want: "dissect",
		},
		{
			name: "main release",
			tag:  "markdown-format--main.5",
			want: "markdown-format",
		},
		{
			name: "project with hyphen",
			tag:  "my-project--pr-1.1",
			want: "my-project",
		},
		{
			name: "single part tag",
			tag:  "project",
			want: "project",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractProjectFromTag(tt.tag)
			if got != tt.want {
				t.Errorf("extractProjectFromTag(%q) = %q, want %q", tt.tag, got, tt.want)
			}
		})
	}
}
