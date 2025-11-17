package pathlang

import (
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// Generators for rapid property testing

// genIdent generates a valid identifier.
func genIdent() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		// Start with letter or underscore
		first := rapid.SampledFrom([]rune{'a', 'b', 'c', 'd', 'e', 'f', '_'}).Draw(t, "first")
		// Continue with letters, digits, underscores, or dashes
		rest := rapid.StringMatching(`[a-z0-9_-]*`).Draw(t, "rest")
		return string(first) + rest
	})
}

// genBareValue generates a bare value (no quotes needed).
func genBareValue() *rapid.Generator[string] {
	// Simple alphanumeric values that don't need quoting
	return rapid.StringMatching(`[a-zA-Z0-9_-]+`)
}

// genQuotedValue generates a value that needs quoting.
func genQuotedValue() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		// Values with spaces, special chars, or requiring escapes
		parts := rapid.SliceOfN(
			rapid.SampledFrom([]string{
				"foo bar",
				"with/slash",
				"with,comma",
				"with[bracket",
				"with]bracket",
				"with\"quote",
				"with\\backslash",
				"with\nnewline",
				"with\ttab",
			}),
			0, 3,
		).Draw(t, "parts")
		return strings.Join(parts, "")
	})
}

// genValue generates either a bare or quoted value.
func genValue() *rapid.Generator[string] {
	return rapid.OneOf(genBareValue(), genQuotedValue())
}

// genOp generates an operator.
func genOp() *rapid.Generator[Op] {
	return rapid.SampledFrom([]Op{OpEq, OpNotEq, OpMatch})
}

// genPredicate generates a predicate.
func genPredicate() *rapid.Generator[Predicate] {
	return rapid.Custom(func(t *rapid.T) Predicate {
		return Predicate{
			Field: genIdent().Draw(t, "field"),
			Op:    genOp().Draw(t, "op"),
			Value: genValue().Draw(t, "value"),
		}
	})
}

// genSegment generates a segment.
func genSegment() *rapid.Generator[Segment] {
	return rapid.Custom(func(t *rapid.T) Segment {
		return Segment{
			Name:       genIdent().Draw(t, "name"),
			Predicates: rapid.SliceOfN(genPredicate(), 0, 3).Draw(t, "predicates"),
		}
	})
}

// genPath generates a random path.
func genPath() *rapid.Generator[*Path] {
	return rapid.Custom(func(t *rapid.T) *Path {
		return &Path{
			Segments: rapid.SliceOfN(genSegment(), 0, 5).Draw(t, "segments"),
		}
	})
}

// Property: Parse → String → Parse round-trip preserves structure
func TestProperty_ParseStringRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random path
		path1 := genPath().Draw(t, "path")

		// Convert to string
		s := path1.String()

		// Parse it back
		path2, err := Parse(s)
		if err != nil {
			t.Fatalf("failed to parse generated path %q: %v", s, err)
		}

		// Should be equal
		if !pathsEqual(path1, path2) {
			t.Fatalf("round-trip failed:\n  original: %+v\n  string:   %q\n  parsed:   %+v",
				path1, s, path2)
		}
	})
}

// Property: String → Parse → String round-trip produces canonical form
func TestProperty_StringParseStringRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random path and convert to string
		path := genPath().Draw(t, "path")
		s1 := path.String()

		// Parse and convert back to string
		parsed, err := Parse(s1)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", s1, err)
		}
		s2 := parsed.String()

		// Strings should be identical (canonical form)
		if s1 != s2 {
			t.Fatalf("string round-trip not canonical:\n  original: %q\n  reparsed: %q", s1, s2)
		}
	})
}

// Property: Root path is empty segments
func TestProperty_RootPathIsEmpty(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		path := &Path{} // Root path

		s := path.String()
		if s != "/" {
			t.Fatalf("root path string should be /, got %q", s)
		}

		parsed, err := Parse(s)
		if err != nil {
			t.Fatalf("failed to parse root path: %v", err)
		}

		if len(parsed.Segments) != 0 {
			t.Fatalf("root path should have zero segments, got %d", len(parsed.Segments))
		}
	})
}

// Property: Segment count is preserved through round-trip
func TestProperty_SegmentCountPreserved(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		path := genPath().Draw(t, "path")
		origCount := len(path.Segments)

		s := path.String()
		parsed, err := Parse(s)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", s, err)
		}

		if len(parsed.Segments) != origCount {
			t.Fatalf("segment count changed: %d -> %d", origCount, len(parsed.Segments))
		}
	})
}

// Property: Predicate count per segment is preserved
func TestProperty_PredicateCountPreserved(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		path := genPath().Draw(t, "path")

		s := path.String()
		parsed, err := Parse(s)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", s, err)
		}

		for i, seg := range path.Segments {
			if len(parsed.Segments[i].Predicates) != len(seg.Predicates) {
				t.Fatalf("predicate count changed in segment %d: %d -> %d",
					i, len(seg.Predicates), len(parsed.Segments[i].Predicates))
			}
		}
	})
}

// Property: Operator types are preserved
func TestProperty_OperatorsPreserved(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		path := genPath().Draw(t, "path")

		s := path.String()
		parsed, err := Parse(s)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", s, err)
		}

		for i, seg := range path.Segments {
			for j, pred := range seg.Predicates {
				if parsed.Segments[i].Predicates[j].Op != pred.Op {
					t.Fatalf("operator changed: %v -> %v",
						pred.Op, parsed.Segments[i].Predicates[j].Op)
				}
			}
		}
	})
}

// Property: Parsing never panics (even on random garbage)
func TestProperty_ParseNeverPanics(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate arbitrary string
		s := rapid.String().Draw(t, "input")

		// Should not panic, even if it returns an error
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("parse panicked on input %q: %v", s, r)
			}
		}()

		_, _ = Parse(s) // Ignore error, just checking for panics
	})
}

// Property: Valid paths always start with /
func TestProperty_ValidPathsStartWithSlash(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		path := genPath().Draw(t, "path")
		s := path.String()

		if !strings.HasPrefix(s, "/") {
			t.Fatalf("valid path should start with /, got %q", s)
		}
	})
}

// Property: Segment names are preserved
func TestProperty_SegmentNamesPreserved(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		path := genPath().Draw(t, "path")

		s := path.String()
		parsed, err := Parse(s)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", s, err)
		}

		for i, seg := range path.Segments {
			if parsed.Segments[i].Name != seg.Name {
				t.Fatalf("segment name changed: %q -> %q", seg.Name, parsed.Segments[i].Name)
			}
		}
	})
}

// Property: Field names are preserved
func TestProperty_FieldNamesPreserved(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		path := genPath().Draw(t, "path")

		s := path.String()
		parsed, err := Parse(s)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", s, err)
		}

		for i, seg := range path.Segments {
			for j, pred := range seg.Predicates {
				if parsed.Segments[i].Predicates[j].Field != pred.Field {
					t.Fatalf("field name changed: %q -> %q",
						pred.Field, parsed.Segments[i].Predicates[j].Field)
				}
			}
		}
	})
}

// Property: Values are preserved (accounting for quoting)
func TestProperty_ValuesPreserved(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		path := genPath().Draw(t, "path")

		s := path.String()
		parsed, err := Parse(s)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", s, err)
		}

		for i, seg := range path.Segments {
			for j, pred := range seg.Predicates {
				expected := pred.Value
				actual := parsed.Segments[i].Predicates[j].Value

				if actual != expected {
					t.Fatalf("value changed: %q -> %q", expected, actual)
				}
			}
		}
	})
}

// Helper function to compare paths structurally
func pathsEqual(a, b *Path) bool {
	if len(a.Segments) != len(b.Segments) {
		return false
	}
	for i := range a.Segments {
		if !segmentsEqual(a.Segments[i], b.Segments[i]) {
			return false
		}
	}
	return true
}

// Helper function to compare segments
func segmentsEqual(a, b Segment) bool {
	if a.Name != b.Name {
		return false
	}
	if len(a.Predicates) != len(b.Predicates) {
		return false
	}
	for i := range a.Predicates {
		if !predicatesEqual(a.Predicates[i], b.Predicates[i]) {
			return false
		}
	}
	return true
}

// Helper function to compare predicates
func predicatesEqual(a, b Predicate) bool {
	return a.Field == b.Field && a.Op == b.Op && a.Value == b.Value
}
