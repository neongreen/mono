package pathlang

import (
	"fmt"
	"strings"
	"unicode"
)

// Parse parses a path string into a Path.
// Returns an error if the input is malformed.
func Parse(input string) (*Path, error) {
	p := &parser{input: input, pos: 0}
	return p.parse()
}

// parser is a hand-written recursive descent parser for pathlang.
type parser struct {
	input string
	pos   int
}

// parse is the main entry point for parsing.
func (p *parser) parse() (*Path, error) {
	// Paths must start with /
	if len(p.input) == 0 {
		return nil, p.errorf("empty path")
	}
	if p.input[0] != '/' {
		return nil, p.errorf("path must start with /")
	}
	p.pos = 1

	// Root path
	if p.pos >= len(p.input) {
		return &Path{}, nil
	}

	// Parse segments
	segments, err := p.parseSegments()
	if err != nil {
		return nil, err
	}

	// Should have consumed all input
	if p.pos < len(p.input) {
		return nil, p.errorf("unexpected character %q", p.input[p.pos])
	}

	return &Path{Segments: segments}, nil
}

// parseSegments parses a sequence of segments separated by /.
func (p *parser) parseSegments() ([]Segment, error) {
	var segments []Segment

	for {
		seg, err := p.parseSegment()
		if err != nil {
			return nil, err
		}
		segments = append(segments, seg)

		// Check for more segments
		if p.pos >= len(p.input) || p.input[p.pos] != '/' {
			break
		}
		p.pos++ // consume /
	}

	return segments, nil
}

// parseSegment parses a single segment: Name Predicates?
func (p *parser) parseSegment() (Segment, error) {
	name, err := p.parseIdent()
	if err != nil {
		return Segment{}, err
	}

	seg := Segment{Name: name}

	// Check for predicates
	if p.pos < len(p.input) && p.input[p.pos] == '[' {
		preds, err := p.parsePredicates()
		if err != nil {
			return Segment{}, err
		}
		seg.Predicates = preds
	}

	return seg, nil
}

// parsePredicates parses a predicate list: "[" PredicateList "]"
func (p *parser) parsePredicates() ([]Predicate, error) {
	if p.pos >= len(p.input) || p.input[p.pos] != '[' {
		return nil, p.errorf("expected '['")
	}
	p.pos++ // consume [

	var preds []Predicate

	for {
		pred, err := p.parsePredicate()
		if err != nil {
			return nil, err
		}
		preds = append(preds, pred)

		// Check for more predicates or end
		if p.pos >= len(p.input) {
			return nil, p.errorf("unclosed predicate list")
		}

		if p.input[p.pos] == ']' {
			p.pos++ // consume ]
			break
		}

		if p.input[p.pos] != ',' {
			return nil, p.errorf("expected ',' or ']' in predicate list")
		}
		p.pos++ // consume ,
	}

	return preds, nil
}

// parsePredicate parses a single predicate: Field Op Value
func (p *parser) parsePredicate() (Predicate, error) {
	field, err := p.parseIdent()
	if err != nil {
		return Predicate{}, fmt.Errorf("in predicate field: %w", err)
	}

	op, err := p.parseOp()
	if err != nil {
		return Predicate{}, err
	}

	value, err := p.parseValue()
	if err != nil {
		return Predicate{}, fmt.Errorf("in predicate value: %w", err)
	}

	return Predicate{Field: field, Op: op, Value: value}, nil
}

// parseIdent parses an identifier: IdentStart IdentCont*
func (p *parser) parseIdent() (string, error) {
	if p.pos >= len(p.input) {
		return "", p.errorf("expected identifier")
	}

	start := p.pos
	r := rune(p.input[p.pos])

	// IdentStart ::= ASCII_LETTER | "_"
	if !isIdentStart(r) {
		return "", p.errorf("expected identifier, got %q", r)
	}
	p.pos++

	// IdentCont ::= ASCII_LETTER | DIGIT | "_" | "-"
	for p.pos < len(p.input) {
		r := rune(p.input[p.pos])
		if !isIdentCont(r) {
			break
		}
		p.pos++
	}

	return p.input[start:p.pos], nil
}

// parseOp parses an operator: "=" | "!=" | "~="
func (p *parser) parseOp() (Op, error) {
	if p.pos >= len(p.input) {
		return 0, p.errorf("expected operator")
	}

	switch p.input[p.pos] {
	case '=':
		p.pos++
		return OpEq, nil
	case '!':
		if p.pos+1 >= len(p.input) || p.input[p.pos+1] != '=' {
			return 0, p.errorf("expected '=' after '!'")
		}
		p.pos += 2
		return OpNotEq, nil
	case '~':
		if p.pos+1 >= len(p.input) || p.input[p.pos+1] != '=' {
			return 0, p.errorf("expected '=' after '~'")
		}
		p.pos += 2
		return OpMatch, nil
	default:
		return 0, p.errorf("expected operator (=, !=, ~=), got %q", p.input[p.pos])
	}
}

// parseValue parses a value: BareValue | QuotedValue
func (p *parser) parseValue() (string, error) {
	if p.pos >= len(p.input) {
		return "", p.errorf("expected value")
	}

	if p.input[p.pos] == '"' {
		return p.parseQuotedValue()
	}
	return p.parseBareValue()
}

// parseBareValue parses an unquoted value: BareChar+
func (p *parser) parseBareValue() (string, error) {
	start := p.pos

	for p.pos < len(p.input) {
		r := rune(p.input[p.pos])
		if !isBareChar(r) {
			break
		}
		p.pos++
	}

	if p.pos == start {
		return "", p.errorf("expected value")
	}

	return p.input[start:p.pos], nil
}

// parseQuotedValue parses a quoted value: `"` QuotedChar* `"`
func (p *parser) parseQuotedValue() (string, error) {
	if p.pos >= len(p.input) || p.input[p.pos] != '"' {
		return "", p.errorf("expected '\"'")
	}
	p.pos++ // consume opening "

	var sb strings.Builder

	for {
		if p.pos >= len(p.input) {
			return "", p.errorf("unclosed quoted string")
		}

		r := rune(p.input[p.pos])

		if r == '"' {
			p.pos++ // consume closing "
			break
		}

		if r == '\\' {
			p.pos++
			if p.pos >= len(p.input) {
				return "", p.errorf("incomplete escape sequence")
			}
			escaped := rune(p.input[p.pos])
			switch escaped {
			case '\\':
				sb.WriteRune('\\')
			case '"':
				sb.WriteRune('"')
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			default:
				return "", p.errorf("invalid escape sequence \\%c", escaped)
			}
			p.pos++
		} else {
			sb.WriteRune(r)
			p.pos++
		}
	}

	return sb.String(), nil
}

// isIdentStart returns true if r can start an identifier.
func isIdentStart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

// isIdentCont returns true if r can continue an identifier.
func isIdentCont(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9') || r == '-'
}

// isBareChar returns true if r can appear in an unquoted value.
// BareChar ::= any char except whitespace, "/", "[", "]", ",", "=", "!", "~", "@", "\"", "'"
func isBareChar(r rune) bool {
	if unicode.IsSpace(r) {
		return false
	}
	switch r {
	case '/', '[', ']', ',', '=', '!', '~', '@', '"', '\'':
		return false
	}
	return true
}

// errorf creates a formatted error with position information.
func (p *parser) errorf(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)

	// Show context around the error position
	start := p.pos - 10
	if start < 0 {
		start = 0
	}
	end := p.pos + 10
	if end > len(p.input) {
		end = len(p.input)
	}

	context := p.input[start:end]
	offset := p.pos - start

	return fmt.Errorf("parse error at position %d: %s\n  context: %q\n  position: %s^",
		p.pos, msg, context, strings.Repeat(" ", offset+10))
}
