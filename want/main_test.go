package main

import (
	"strings"
	"testing"
)

func TestParseMonoProjectVersion(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantProject    string
		wantVersion    string
		wantListFormat bool
	}{
		{
			name:           "project with version",
			input:          "printpdf@main.1",
			wantProject:    "printpdf",
			wantVersion:    "main.1",
			wantListFormat: false,
		},
		{
			name:           "project without version",
			input:          "printpdf",
			wantProject:    "printpdf",
			wantVersion:    "",
			wantListFormat: true,
		},
		{
			name:           "project with pr version",
			input:          "dissect@pr-42.1",
			wantProject:    "dissect",
			wantVersion:    "pr-42.1",
			wantListFormat: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.Split(tt.input, "@")

			gotProject := parts[0]
			if gotProject != tt.wantProject {
				t.Errorf("project = %v, want %v", gotProject, tt.wantProject)
			}

			if len(parts) == 1 {
				if !tt.wantListFormat {
					t.Errorf("expected list format but got version format")
				}
				return
			}

			if len(parts) == 2 {
				gotVersion := parts[1]
				if gotVersion != tt.wantVersion {
					t.Errorf("version = %v, want %v", gotVersion, tt.wantVersion)
				}
			}
		})
	}
}

func TestFormatMonoTag(t *testing.T) {
	tests := []struct {
		name    string
		project string
		version string
		want    string
	}{
		{
			name:    "main release",
			project: "printpdf",
			version: "main.1",
			want:    "printpdf--main.1",
		},
		{
			name:    "pr release",
			project: "dissect",
			version: "pr-42.1",
			want:    "dissect--pr-42.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.project + "--" + tt.version
			if got != tt.want {
				t.Errorf("tag = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractVersionFromTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		project string
		want    string
	}{
		{
			name:    "main release",
			tag:     "printpdf--main.1",
			project: "printpdf",
			want:    "main.1",
		},
		{
			name:    "pr release",
			tag:     "dissect--pr-42.1",
			project: "dissect",
			want:    "pr-42.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix := tt.project + "--"
			got := strings.TrimPrefix(tt.tag, prefix)
			if got != tt.want {
				t.Errorf("version = %v, want %v", got, tt.want)
			}
		})
	}
}
