package parse

import (
	"testing"
)

func TestParseLenses(t *testing.T) {
	content := `# Lenses

@lens[default]
This is the default lens.
It has multiple lines.

@lens[pedantic]
This is the pedantic lens.
Also multiple lines.
`

	lenses, err := ParseLenses(content, "test.md")
	if err != nil {
		t.Fatalf("ParseLenses() error = %v", err)
	}

	if len(lenses) != 2 {
		t.Fatalf("got %d lenses, want 2", len(lenses))
	}

	if _, ok := lenses["default"]; !ok {
		t.Error("missing default lens")
	}

	if _, ok := lenses["pedantic"]; !ok {
		t.Error("missing pedantic lens")
	}
}
