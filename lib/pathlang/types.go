package pathlang

import (
	"fmt"
	"strings"
)

// Op represents a predicate comparison operator.
type Op int

const (
	// OpEq represents the equality operator (=).
	OpEq Op = iota
	// OpNotEq represents the inequality operator (!=).
	OpNotEq
	// OpMatch represents the pattern match operator (~=).
	// Semantics are defined by the resolver (typically contains or glob).
	OpMatch
)

// String returns the string representation of an operator.
func (o Op) String() string {
	switch o {
	case OpEq:
		return "="
	case OpNotEq:
		return "!="
	case OpMatch:
		return "~="
	default:
		return fmt.Sprintf("Op(%d)", o)
	}
}

// Predicate represents a single field comparison within a segment.
// All predicates in a segment are implicitly ANDed together.
type Predicate struct {
	// Field is the name of the field to test.
	Field string
	// Op is the comparison operator.
	Op Op
	// Value is the value to compare against (always a string lexically).
	Value string
}

// String returns the canonical string representation of a predicate.
func (p Predicate) String() string {
	return p.Field + p.Op.String() + quoteValue(p.Value)
}

// Segment represents a single path segment with optional predicates.
type Segment struct {
	// Name is the segment name (e.g., "projects", "tasks").
	Name string
	// Predicates are the filters applied at this segment.
	// All predicates must be satisfied (implicit AND).
	Predicates []Predicate
}

// String returns the canonical string representation of a segment.
func (s Segment) String() string {
	var sb strings.Builder
	sb.WriteString(s.Name)
	if len(s.Predicates) > 0 {
		sb.WriteString("[")
		for i, p := range s.Predicates {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(p.String())
		}
		sb.WriteString("]")
	}
	return sb.String()
}

// Path represents a complete path expression.
// Paths are always absolute (start with /).
type Path struct {
	// Segments are the path components.
	// Empty for root path (/).
	Segments []Segment
}

// String returns the canonical string representation of a path.
func (p *Path) String() string {
	if len(p.Segments) == 0 {
		return "/"
	}
	var sb strings.Builder
	for _, seg := range p.Segments {
		sb.WriteString("/")
		sb.WriteString(seg.String())
	}
	return sb.String()
}

// quoteValue returns the canonical quoted form of a value.
// Values that need quoting are always double-quoted.
func quoteValue(s string) string {
	// Check if the value needs quoting
	if needsQuoting(s) {
		return quote(s)
	}
	return s
}

// needsQuoting returns true if a value needs to be quoted.
func needsQuoting(s string) bool {
	if len(s) == 0 {
		return true
	}
	for _, r := range s {
		// BareChar ::= any char except whitespace, "/", "[", "]", ",", "=", "!", "~", "@", "\"", "'", "\\"
		switch r {
		case ' ', '\t', '\n', '\r', '/', '[', ']', ',', '=', '!', '~', '@', '"', '\'', '\\':
			return true
		}
	}
	return false
}

// quote returns a double-quoted string with proper escaping.
func quote(s string) string {
	var sb strings.Builder
	sb.WriteString(`"`)
	for _, r := range s {
		switch r {
		case '\\':
			sb.WriteString(`\\`)
		case '"':
			sb.WriteString(`\"`)
		case '\n':
			sb.WriteString(`\n`)
		case '\t':
			sb.WriteString(`\t`)
		default:
			sb.WriteRune(r)
		}
	}
	sb.WriteString(`"`)
	return sb.String()
}
