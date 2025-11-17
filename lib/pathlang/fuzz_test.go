package pathlang

import "testing"

// FuzzParse tests that Parse never panics on arbitrary input.
// This is a native Go fuzz test (Go 1.18+).
//
// Run with: go test -fuzz=FuzzParse
func FuzzParse(f *testing.F) {
	// Seed corpus with known valid inputs
	seeds := []string{
		"/",
		"/projects",
		"/projects/tasks",
		"/projects[name=foo]",
		"/projects[name=foo]/tasks[status=open]",
		"/tasks[status=open,assignee=me]",
		`/projects[name="foo bar"]`,
		`/projects[description="with\"quotes"]`,
		"/tasks[id=42]",
		"/tasks[status!=closed]",
		"/tasks[desc~=bug]",
		// Edge cases
		"",
		"projects",
		"/projects[",
		"/projects]",
		"/projects[name",
		"/projects[name=",
		"/projects[name=]",
		"/projects[=value]",
		"/projects[name==value]",
		`/projects[name="unclosed`,
		`/projects[name="\x"]`,
		"/projects//tasks",
		"/projects/",
		// Invalid but interesting
		"///",
		"/[[[",
		"/]]]",
		"/ /",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Parse should never panic, even on invalid input
		_, _ = Parse(input)
	})
}

// FuzzParseRoundTrip tests that valid paths round-trip correctly.
func FuzzParseRoundTrip(f *testing.F) {
	// Seed with valid paths only
	seeds := []string{
		"/",
		"/projects",
		"/projects/tasks",
		"/projects[name=foo]",
		"/projects[name=foo]/tasks[status=open,assignee=me]",
		`/projects[name="foo bar"]`,
		`/projects[description="line1\nline2\ttab"]`,
		"/tasks[id!=42]",
		"/tasks[desc~=bug]",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Try to parse
		path1, err := Parse(input)
		if err != nil {
			// Invalid input is fine, just skip
			return
		}

		// Convert to string
		s := path1.String()

		// Parse again
		path2, err := Parse(s)
		if err != nil {
			t.Fatalf("failed to parse generated path %q: %v", s, err)
		}

		// Should be equal
		if !pathsEqual(path1, path2) {
			t.Errorf("round-trip failed:\n  original: %+v\n  string:   %q\n  parsed:   %+v",
				path1, s, path2)
		}
	})
}
