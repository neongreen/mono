package validator

import (
	"fmt"
	"strings"
	"testing"
)

// TestValidateScript tests the ValidateScript function with various shell script patterns
func TestValidateScript(t *testing.T) {
	tests := []struct {
		name       string
		script     string
		wantCount  int
		wantInLine []int // expected line numbers with violations
	}{
		{
			name:       "simple cd outside subshell",
			script:     "cd /tmp",
			wantCount:  1,
			wantInLine: []int{1},
		},
		{
			name:       "cd inside subshell - allowed",
			script:     "(cd /tmp && ls)",
			wantCount:  0,
			wantInLine: []int{},
		},
		{
			name:       "multiple cd outside subshell",
			script:     "cd /tmp\ncd /home\ncd /var",
			wantCount:  3,
			wantInLine: []int{1, 2, 3},
		},
		{
			name:       "mixed cd inside and outside subshell",
			script:     "cd /tmp\n(cd /home && ls)\ncd /var",
			wantCount:  2,
			wantInLine: []int{1, 3},
		},
		{
			name:       "cd with arguments",
			script:     "cd /tmp/test/dir",
			wantCount:  1,
			wantInLine: []int{1},
		},
		{
			name:       "cd in command substitution subshell",
			script:     "result=$(cd /tmp && pwd)",
			wantCount:  0,
			wantInLine: []int{},
		},
		{
			name:       "cd in pipeline subshell",
			script:     "(cd /tmp && find .) | grep test",
			wantCount:  0,
			wantInLine: []int{},
		},
		{
			name:       "nested subshells with cd",
			script:     "(cd /tmp && (cd /home && ls))",
			wantCount:  0,
			wantInLine: []int{},
		},
		{
			name:       "cd with conditional - outside subshell",
			script:     "if [ -d /tmp ]; then cd /tmp; fi",
			wantCount:  1,
			wantInLine: []int{1},
		},
		{
			name:       "cd in for loop - outside subshell",
			script:     "for dir in /tmp /home; do cd $dir; done",
			wantCount:  1,
			wantInLine: []int{1},
		},
		{
			name:       "cd in while loop - outside subshell",
			script:     "while read dir; do cd $dir; done",
			wantCount:  1,
			wantInLine: []int{1},
		},
		{
			name:       "cd in function definition - outside subshell",
			script:     "myfunc() { cd /tmp; }",
			wantCount:  1,
			wantInLine: []int{1},
		},
		{
			name:       "no cd command",
			script:     "ls /tmp\necho hello\npwd",
			wantCount:  0,
			wantInLine: []int{},
		},
		{
			name:       "command containing cd but not cd itself",
			script:     "abcd /tmp",
			wantCount:  0,
			wantInLine: []int{},
		},
		{
			name:       "cd as part of string",
			script:     "echo 'cd /tmp'",
			wantCount:  0,
			wantInLine: []int{},
		},
		{
			name:       "cd in background job subshell",
			script:     "(cd /tmp && ls) &",
			wantCount:  0,
			wantInLine: []int{},
		},
		{
			name:       "cd with variable",
			script:     "cd $HOME",
			wantCount:  1,
			wantInLine: []int{1},
		},
		{
			name:       "cd with tilde",
			script:     "cd ~",
			wantCount:  1,
			wantInLine: []int{1},
		},
		{
			name:       "cd without arguments (go to home)",
			script:     "cd",
			wantCount:  1,
			wantInLine: []int{1},
		},
		{
			name:       "multiple commands on one line with cd",
			script:     "echo start && cd /tmp && echo done",
			wantCount:  1,
			wantInLine: []int{1},
		},
		{
			name:       "cd in case statement",
			script:     "case $var in x) cd /tmp;; esac",
			wantCount:  1,
			wantInLine: []int{1},
		},
		{
			name:       "cd with -P flag",
			script:     "cd -P /tmp",
			wantCount:  1,
			wantInLine: []int{1},
		},
		{
			name: "multiline script with mixed violations",
			script: `#!/bin/bash
set -e
cd /tmp
(cd /home && ls)
mkdir -p test
cd test
(cd /var && pwd)`,
			wantCount:  2,
			wantInLine: []int{3, 6},
		},
		{
			name: "cd in here document - should not detect",
			script: `cat <<EOF
cd /tmp
EOF`,
			wantCount:  0,
			wantInLine: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			violations, err := v.ValidateScript(strings.NewReader(tt.script))
			if err != nil {
				t.Fatalf("ValidateScript() error = %v, want no error\nScript:\n%s", err, tt.script)
			}

			if len(violations) != tt.wantCount {
				t.Errorf("ValidateScript() got %d violations, want %d\nViolations: %v\nScript:\n%s",
					len(violations), tt.wantCount, violations, tt.script)
			}

			// Check line numbers if specified
			if len(tt.wantInLine) > 0 {
				for _, wantLine := range tt.wantInLine {
					found := false
					expectedLineStr := fmt.Sprintf("Line %d", wantLine)
					for _, v := range violations {
						if strings.Contains(v, expectedLineStr) {
							found = true
							break
						}
					}
					if !found && tt.wantCount > 0 {
						t.Errorf("Expected violation on line %d but not found.\nViolations: %v\nScript:\n%s",
							wantLine, violations, tt.script)
					}
				}
			}
		})
	}
}

func TestFormatViolations(t *testing.T) {
	tests := []struct {
		name       string
		violations []string
		wantSubstr []string
	}{
		{
			name:       "single violation",
			violations: []string{"Line 1: 'cd' command found outside subshell"},
			wantSubstr: []string{"Line 1", "cd", "subshell", "Good: (cd"},
		},
		{
			name:       "multiple violations",
			violations: []string{"Line 1: 'cd' command found outside subshell", "Line 3: 'cd' command found outside subshell"},
			wantSubstr: []string{"Line 1", "Line 3", "cd", "subshell"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatViolations(tt.violations)
			for _, substr := range tt.wantSubstr {
				if !strings.Contains(got, substr) {
					t.Errorf("FormatViolations() missing substring %q\nGot: %s", substr, got)
				}
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		script    string
		wantError bool
	}{
		{
			name:      "empty script",
			script:    "",
			wantError: false,
		},
		{
			name:      "whitespace only",
			script:    "   \n\t\n  ",
			wantError: false,
		},
		{
			name:      "comment only",
			script:    "# This is a comment",
			wantError: false,
		},
		{
			name:      "invalid syntax - unclosed quote",
			script:    `echo "unclosed`,
			wantError: true,
		},
		{
			name:      "invalid syntax - unclosed subshell",
			script:    "(cd /tmp",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			_, err := v.ValidateScript(strings.NewReader(tt.script))

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none for script: %s", tt.script)
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v for script: %s", err, tt.script)
			}
		})
	}
}

func TestComplexNesting(t *testing.T) {
	tests := []struct {
		name      string
		script    string
		wantCount int
	}{
		{
			name:      "deeply nested subshells",
			script:    "(cd /tmp && (cd /home && ls))",
			wantCount: 0,
		},
		{
			name:      "mixed nesting with violations",
			script:    "cd /a\n(cd /b && ls)\ncd /c",
			wantCount: 2,
		},
		{
			name:      "subshell in command substitution",
			script:    "x=$(cd /tmp && pwd)",
			wantCount: 0,
		},
		{
			name:      "multiple command substitutions",
			script:    "a=$(cd /tmp && pwd); b=$(cd /home && pwd)",
			wantCount: 0,
		},
		{
			name:      "cd outside with subshell inside conditional",
			script:    "cd /tmp\nif true; then (cd /home && ls); fi",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			violations, err := v.ValidateScript(strings.NewReader(tt.script))
			if err != nil {
				t.Fatalf("ValidateScript() error = %v", err)
			}

			if len(violations) != tt.wantCount {
				t.Errorf("ValidateScript() got %d violations, want %d\nViolations: %v\nScript:\n%s",
					len(violations), tt.wantCount, violations, tt.script)
			}
		})
	}
}
