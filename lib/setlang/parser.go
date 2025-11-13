package setlang

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Parser is a parser for the set language.
type Parser struct {
	parser *participle.Parser[Expr]
}

// NewParser creates a new parser for the set language.
func NewParser() (*Parser, error) {
	// Define a custom lexer that handles identifiers, strings, operators, and whitespace
	lex := lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Whitespace", Pattern: `[ \t\r\n]+`},
		{Name: "String", Pattern: `"(?:[^"\\]|\\.)*"`},
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
		{Name: "Punct", Pattern: `[|&\-(),]`},
	})

	parser, err := participle.Build[Expr](
		participle.Lexer(lex),
		participle.Elide("Whitespace"),
		participle.UseLookahead(2),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build parser: %w", err)
	}

	return &Parser{parser: parser}, nil
}

// Parse parses an input string into an AST.
func (p *Parser) Parse(input string) (*Expr, error) {
	expr, err := p.parser.ParseString("", input)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	return expr, nil
}

// MustParse parses an input string and panics on error.
// This is useful for testing.
func (p *Parser) MustParse(input string) *Expr {
	expr, err := p.Parse(input)
	if err != nil {
		panic(err)
	}
	return expr
}

// Parse is a convenience function that creates a new parser and parses the input.
func Parse(input string) (*Expr, error) {
	p, err := NewParser()
	if err != nil {
		return nil, err
	}
	return p.Parse(input)
}

// MustParse is a convenience function that creates a new parser and parses the input,
// panicking on error.
func MustParse(input string) *Expr {
	expr, err := Parse(input)
	if err != nil {
		panic(err)
	}
	return expr
}

// String returns a string representation of an expression.
// This is useful for debugging.
func (e *Expr) String() string {
	if e.Union == nil {
		return ""
	}
	return e.Union.String()
}

func (u *UnionExpr) String() string {
	var sb strings.Builder
	sb.WriteString(u.Left.String())
	for _, tail := range u.Right {
		sb.WriteString(" | ")
		sb.WriteString(tail.Right.String())
	}
	return sb.String()
}

func (i *IntersectExpr) String() string {
	var sb strings.Builder
	sb.WriteString(i.Left.String())
	for _, tail := range i.Right {
		sb.WriteString(" & ")
		sb.WriteString(tail.Right.String())
	}
	return sb.String()
}

func (d *DiffExpr) String() string {
	var sb strings.Builder
	sb.WriteString(d.Left.String())
	for _, tail := range d.Right {
		sb.WriteString(" - ")
		sb.WriteString(tail.Right.String())
	}
	return sb.String()
}

func (p *Primary) String() string {
	if p.FuncCall != nil {
		return p.FuncCall.String()
	}
	if p.Ident != nil {
		return *p.Ident
	}
	if p.SubExpr != nil {
		return "(" + p.SubExpr.String() + ")"
	}
	return ""
}

func (f *FuncCall) String() string {
	var sb strings.Builder
	sb.WriteString(f.Name)
	sb.WriteString("(")
	for i, arg := range f.Args {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(arg.String())
	}
	sb.WriteString(")")
	return sb.String()
}

func (a *Arg) String() string {
	if a.Expr != nil {
		return a.Expr.String()
	}
	if a.StrLit != nil {
		return *a.StrLit
	}
	if a.Ident != nil {
		return *a.Ident
	}
	return ""
}
