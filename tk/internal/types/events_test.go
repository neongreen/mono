package types

import (
	"strings"
	"testing"
)

// TestValidateProjectName tests the project name validation function
func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		wantErr     bool
		wantSuggest bool // whether we expect a suggestion in the error
	}{
		// Valid names
		{
			name:        "valid simple name",
			projectName: "foo",
			wantErr:     false,
		},
		{
			name:        "valid with dash",
			projectName: "foo-bar",
			wantErr:     false,
		},
		{
			name:        "valid with multiple dashes",
			projectName: "foo-bar-baz",
			wantErr:     false,
		},
		{
			name:        "valid single letter",
			projectName: "a",
			wantErr:     false,
		},
		{
			name:        "valid long name",
			projectName: "my-very-long-project-name-with-many-parts",
			wantErr:     false,
		},

		// Invalid - empty
		{
			name:        "empty name",
			projectName: "",
			wantErr:     true,
			wantSuggest: false,
		},

		// Invalid - uppercase
		{
			name:        "uppercase first letter",
			projectName: "Foo",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "uppercase middle letter",
			projectName: "fooBar",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "all uppercase",
			projectName: "FOO",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "uppercase with dash",
			projectName: "Foo-Bar",
			wantErr:     true,
			wantSuggest: true,
		},

		// Invalid - leading/trailing dashes
		{
			name:        "leading dash",
			projectName: "-foo",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "trailing dash",
			projectName: "foo-",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "leading and trailing dash",
			projectName: "-foo-",
			wantErr:     true,
			wantSuggest: true,
		},

		// Invalid - consecutive dashes
		{
			name:        "consecutive dashes",
			projectName: "foo--bar",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "triple consecutive dashes",
			projectName: "foo---bar",
			wantErr:     true,
			wantSuggest: true,
		},

		// Invalid - digits
		{
			name:        "trailing digit",
			projectName: "foo-bar-123",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "leading digit",
			projectName: "123-foo",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "digit in middle",
			projectName: "foo-1-bar",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "only digits",
			projectName: "123",
			wantErr:     true,
			wantSuggest: false,
		},

		// Invalid - other characters
		{
			name:        "underscore",
			projectName: "foo_bar",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "space",
			projectName: "foo bar",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "special characters",
			projectName: "foo@bar",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "dot",
			projectName: "foo.bar",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "slash",
			projectName: "foo/bar",
			wantErr:     true,
			wantSuggest: true,
		},

		// Edge cases - combinations
		{
			name:        "uppercase and underscore",
			projectName: "Foo_Bar",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "uppercase and digit",
			projectName: "Foo123",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "leading dash and uppercase",
			projectName: "-Foo",
			wantErr:     true,
			wantSuggest: true,
		},
		{
			name:        "only dash",
			projectName: "-",
			wantErr:     true,
			wantSuggest: false,
		},
		{
			name:        "only dashes",
			projectName: "---",
			wantErr:     true,
			wantSuggest: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectName(tt.projectName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectName(%q) error = %v, wantErr %v", tt.projectName, err, tt.wantErr)
				return
			}

			// If we expect an error with a suggestion, check for "Try:" in the error message
			if tt.wantErr && tt.wantSuggest && err != nil {
				if !strings.Contains(err.Error(), "Try:") {
					t.Errorf("ValidateProjectName(%q) error = %v, expected error with suggestion (containing 'Try:')", tt.projectName, err)
				}
			}
		})
	}
}

// TestValidateProjectName_Suggestions tests that the suggestions are sensible
func TestValidateProjectName_Suggestions(t *testing.T) {
	tests := []struct {
		name            string
		projectName     string
		wantSuggestion  string
	}{
		{
			name:           "uppercase to lowercase",
			projectName:    "Foo",
			wantSuggestion: "foo",
		},
		{
			name:           "uppercase with dash",
			projectName:    "Foo-Bar",
			wantSuggestion: "foo-bar",
		},
		{
			name:           "trailing dash removed",
			projectName:    "foo-bar-",
			wantSuggestion: "foo-bar",
		},
		{
			name:           "leading dash removed",
			projectName:    "-foo-bar",
			wantSuggestion: "foo-bar",
		},
		{
			name:           "consecutive dashes to single",
			projectName:    "foo--bar",
			wantSuggestion: "foo-bar",
		},
		{
			name:           "underscore removed",
			projectName:    "foo_bar",
			wantSuggestion: "foobar",
		},
		{
			name:           "digits removed",
			projectName:    "foo-123",
			wantSuggestion: "foo-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectName(tt.projectName)
			if err == nil {
				t.Fatalf("ValidateProjectName(%q) expected error but got nil", tt.projectName)
			}

			if !strings.Contains(err.Error(), tt.wantSuggestion) {
				t.Errorf("ValidateProjectName(%q) error = %v, expected error containing suggestion %q", tt.projectName, err, tt.wantSuggestion)
			}
		})
	}
}
