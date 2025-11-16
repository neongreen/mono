package pathlang

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// AST types for participle parsing
// These are separate from the domain types to keep parsing concerns separate

// pathAST is the internal AST representation used by participle.
type pathAST struct {
	Segments []*segmentAST `"/" @@? ( "/" @@ )*`
}

// segmentAST represents a parsed segment.
type segmentAST struct {
	Name       string         `@(BareValue | Ident)`
	Predicates *predicatesAST `@@?`
}

// predicatesAST represents a list of predicates in brackets.
type predicatesAST struct {
	Predicates []*predicateAST `"[" @@ ( "," @@ )* "]"`
}

// predicateAST represents a single predicate.
type predicateAST struct {
	Field string `@(BareValue | Ident)`
	Op    string `@( "!=" | "~=" | "=" )`
	Value string `@( String | BareValue | Ident )`
}

// Parser is a parser for the pathlang language.
type Parser struct {
	parser *participle.Parser[pathAST]
}

// NewParser creates a new parser for pathlang.
func NewParser() (*Parser, error) {
	// Define a custom lexer that handles identifiers, strings, operators, and punctuation
	// Order matters: more specific patterns first
	lex := lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Whitespace", Pattern: `[ \t\r\n]+`},
		{Name: "String", Pattern: `"(?:[^"\\]|\\.)*"`},
		{Name: "NotEq", Pattern: `!=`},
		{Name: "Match", Pattern: `~=`},
		{Name: "Eq", Pattern: `=`},
		{Name: "Punct", Pattern: `[/\[\],]`},
		// BareValue must come before Ident so values like "foo.bar" are captured as a single token
		// Otherwise "foo.bar" would be tokenized as Ident("foo") + BareValue(".bar")
		{Name: "BareValue", Pattern: `[^ \t\n\r/\[\],=!~@"']+`},
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_-]*`},
	})

	parser, err := participle.Build[pathAST](
		participle.Lexer(lex),
		participle.Elide("Whitespace"),
		participle.UseLookahead(2),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build parser: %w", err)
	}

	return &Parser{parser: parser}, nil
}

// Parse parses an input string into a Path.
func (p *Parser) Parse(input string) (*Path, error) {
	if input == "" {
		return nil, fmt.Errorf("empty path")
	}

	ast, err := p.parser.ParseString("", input)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return astToPath(ast)
}

// MustParse parses an input string and panics on error.
// This is useful for testing.
func (p *Parser) MustParse(input string) *Path {
	path, err := p.Parse(input)
	if err != nil {
		panic(err)
	}
	return path
}

// Parse is a convenience function that creates a new parser and parses the input.
func Parse(input string) (*Path, error) {
	p, err := NewParser()
	if err != nil {
		return nil, err
	}
	return p.Parse(input)
}

// MustParse is a convenience function that creates a new parser and parses the input,
// panicking on error.
func MustParse(input string) *Path {
	path, err := Parse(input)
	if err != nil {
		panic(err)
	}
	return path
}

// astToPath converts the participle AST to our domain Path type.
func astToPath(ast *pathAST) (*Path, error) {
	segments := make([]Segment, len(ast.Segments))

	for i, segAST := range ast.Segments {
		// Validate that segment name is a valid identifier
		if !isValidIdentifier(segAST.Name) {
			return nil, fmt.Errorf("invalid segment name %q: must start with letter or underscore and contain only alphanumeric, underscore, or dash characters", segAST.Name)
		}

		seg := Segment{
			Name: segAST.Name,
		}

		if segAST.Predicates != nil {
			preds := make([]Predicate, len(segAST.Predicates.Predicates))
			for j, predAST := range segAST.Predicates.Predicates {
				// Validate that field name is a valid identifier
				if !isValidIdentifier(predAST.Field) {
					return nil, fmt.Errorf("invalid field name %q in predicate: must start with letter or underscore and contain only alphanumeric, underscore, or dash characters", predAST.Field)
				}

				op, err := parseOp(predAST.Op)
				if err != nil {
					return nil, err
				}

				value, err := unescapeValue(predAST.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid value in predicate %s: %w", predAST.Field, err)
				}

				preds[j] = Predicate{
					Field: predAST.Field,
					Op:    op,
					Value: value,
				}
			}
			seg.Predicates = preds
		}

		segments[i] = seg
	}

	return &Path{Segments: segments}, nil
}

// parseOp converts the operator string to an Op value.
func parseOp(op string) (Op, error) {
	switch op {
	case "=":
		return OpEq, nil
	case "!=":
		return OpNotEq, nil
	case "~=":
		return OpMatch, nil
	default:
		return 0, fmt.Errorf("unknown operator: %s", op)
	}
}

// isValidIdentifier checks if a string is a valid identifier.
// Valid identifiers start with a letter or underscore and contain only
// alphanumeric characters, underscores, or dashes.
func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	// First character must be letter or underscore
	first := rune(s[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}
	// Remaining characters must be alphanumeric, underscore, or dash
	for i := 1; i < len(s); i++ {
		c := rune(s[i])
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

// unescapeValue processes a value, removing quotes and handling escape sequences if needed.
func unescapeValue(s string) (string, error) {
	// If it's a quoted string, remove quotes and unescape
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]

		// Process escape sequences
		var result strings.Builder
		i := 0
		for i < len(s) {
			if s[i] == '\\' && i+1 < len(s) {
				switch s[i+1] {
				case '\\':
					result.WriteRune('\\')
				case '"':
					result.WriteRune('"')
				case 'n':
					result.WriteRune('\n')
				case 't':
					result.WriteRune('\t')
				default:
					return "", fmt.Errorf("invalid escape sequence \\%c", s[i+1])
				}
				i += 2
			} else {
				result.WriteByte(s[i])
				i++
			}
		}
		return result.String(), nil
	}

	// Bare value, return as-is
	return s, nil
}
