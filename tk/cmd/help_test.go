package cmd

import (
	"strings"
	"testing"
)

// TestSeeAlsoValidation ensures all "See Also" references point to
// actual commands that exist in the command tree.
//
// This test will fail if:
// - A command in the registry doesn't exist
// - A referenced command doesn't exist
// - There are typos in command names
func TestSeeAlsoValidation(t *testing.T) {
	if err := ValidateSeeAlso(RootCmd); err != nil {
		t.Fatalf("See Also validation failed:\n%v", err)
	}
}

// TestSeeAlsoRegistry checks basic invariants of the registry.
func TestSeeAlsoRegistry(t *testing.T) {
	// Ensure registry is not empty
	if len(seeAlsoRegistry) == 0 {
		t.Error("seeAlsoRegistry is empty - did you populate it?")
	}

	// Ensure no command references itself
	for cmd, related := range seeAlsoRegistry {
		for _, rel := range related {
			if cmd == rel {
				t.Errorf("command %q references itself in See Also", cmd)
			}
		}
	}

	// Ensure no empty reference lists
	for cmd, related := range seeAlsoRegistry {
		if len(related) == 0 {
			t.Errorf("command %q has empty See Also list", cmd)
		}
	}
}

// TestSeeAlsoFormatting checks that the SeeAlso formatter produces
// the expected output format.
func TestSeeAlsoFormatting(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		want     string
	}{
		{
			name:     "empty",
			commands: []string{},
			want:     "",
		},
		{
			name:     "single command",
			commands: []string{"show"},
			want:     "\n\nSee Also:\n  tk show",
		},
		{
			name:     "multiple commands",
			commands: []string{"show", "edit", "mark"},
			want:     "\n\nSee Also:\n  tk show\n  tk edit\n  tk mark",
		},
		{
			name:     "subcommand",
			commands: []string{"relate add"},
			want:     "\n\nSee Also:\n  tk relate add",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SeeAlso(tt.commands...)
			if got != tt.want {
				t.Errorf("SeeAlso() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSeeAlsoWithDescriptions checks that the description formatter
// correctly pulls descriptions from Cobra commands.
func TestSeeAlsoWithDescriptions(t *testing.T) {
	// Apply see also to populate all commands
	ApplySeeAlso(RootCmd)

	tests := []struct {
		name     string
		commands []string
		contains []string // strings that should appear in output
	}{
		{
			name:     "empty",
			commands: []string{},
			contains: []string{},
		},
		{
			name:     "commands with descriptions",
			commands: []string{"show", "edit"},
			contains: []string{"tk show", "tk edit", "Show task details", "Edit task fields"},
		},
		{
			name:     "subcommand",
			commands: []string{"relate-add"},
			contains: []string{"tk relate-add", "Add a relation between two tasks"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SeeAlsoWithDescriptions(RootCmd, tt.commands...)

			// Check that all expected strings are present
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("SeeAlsoWithDescriptions() output missing %q\nGot: %s", want, got)
				}
			}
		})
	}
}
