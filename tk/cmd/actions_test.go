package cmd

import (
	"testing"

	"github.com/neongreen/mono/lib/pathlang"
)

// TestActionParsing tests that actions are properly parsed from path strings
func TestActionParsing(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantAction string
		wantArgs   []string
	}{
		{
			name:       "simple action",
			input:      "/me-1 @status",
			wantAction: "status",
			wantArgs:   nil,
		},
		{
			name:       "action with single arg",
			input:      "/me-1/notes @add hello",
			wantAction: "add",
			wantArgs:   []string{"hello"},
		},
		{
			name:       "action with multiple args",
			input:      "/me-1/notes @add hello world",
			wantAction: "add",
			wantArgs:   []string{"hello", "world"},
		},
		{
			name:       "action with quoted arg",
			input:      `/me-1/notes @add "hello world"`,
			wantAction: "add",
			wantArgs:   []string{"hello world"},
		},
		{
			name:       "no action",
			input:      "/me-1/notes",
			wantAction: "",
			wantArgs:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := pathlang.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if path.Action != tt.wantAction {
				t.Errorf("Action = %q, want %q", path.Action, tt.wantAction)
			}

			if len(path.ActionArgs) != len(tt.wantArgs) {
				t.Errorf("ActionArgs length = %d, want %d", len(path.ActionArgs), len(tt.wantArgs))
			} else {
				for i, arg := range path.ActionArgs {
					if arg != tt.wantArgs[i] {
						t.Errorf("ActionArgs[%d] = %q, want %q", i, arg, tt.wantArgs[i])
					}
				}
			}
		})
	}
}
