package setlang

import (
	"testing"
)

func TestParserBasic(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple identifier",
			input: "foo",
			want:  "foo",
		},
		{
			name:  "function call no args",
			input: "all()",
			want:  "all()",
		},
		{
			name:  "function call with one arg",
			input: "status(done)",
			want:  "status(done)",
		},
		{
			name:  "function call with multiple args",
			input: "func(a, b, c)",
			want:  "func(a, b, c)",
		},
		{
			name:  "function call with string arg",
			input: `title("hello world")`,
			want:  `title("hello world")`,
		},
		{
			name:  "union operation",
			input: "a | b",
			want:  "a | b",
		},
		{
			name:  "intersect operation",
			input: "a & b",
			want:  "a & b",
		},
		{
			name:  "diff operation",
			input: "a - b",
			want:  "a - b",
		},
		{
			name:  "parentheses",
			input: "(a | b)",
			want:  "(a | b)",
		},
		{
			name:  "complex expression",
			input: "a | b & c",
			want:  "a | b & c",
		},
		{
			name:  "precedence test",
			input: "a | b & c - d",
			want:  "a | b & c - d",
		},
		{
			name:  "parentheses override precedence",
			input: "(a | b) & c",
			want:  "(a | b) & c",
		},
		{
			name:  "nested parentheses",
			input: "((a | b) & c) - d",
			want:  "((a | b) & c) - d",
		},
		{
			name:  "function with expression arg",
			input: "subtasks(a | b)",
			want:  "subtasks(a | b)",
		},
		{
			name:  "complex nested expression",
			input: "status(open) & (label(bug) | label(feature)) - archived",
			want:  "status(open) & (label(bug) | label(feature)) - archived",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			got := expr.String()
			if got != tt.want {
				t.Errorf("Parse() output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParserErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unclosed parenthesis",
			input: "(a | b",
		},
		{
			name:  "unclosed function call",
			input: "func(a, b",
		},
		{
			name:  "empty input",
			input: "",
		},
		{
			name:  "trailing operator",
			input: "a |",
		},
		{
			name:  "leading operator",
			input: "| a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil {
				t.Error("Parse() expected error, got nil")
			}
		})
	}
}

func TestParserPrecedence(t *testing.T) {
	// Test that operators have correct precedence: | < & < -
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	// a | b & c should parse as a | (b & c)
	expr1 := parser.MustParse("a | b & c")
	union := expr1.Union
	if len(union.Right) != 1 {
		t.Error("expected one union operation")
	}
	// The right side should have an intersection
	intersect := union.Right[0].Right
	if len(intersect.Right) != 1 {
		t.Error("expected intersection on right side of union")
	}

	// a & b - c should parse as a & (b - c)
	expr2 := parser.MustParse("a & b - c")
	intersect2 := expr2.Union.Left
	if len(intersect2.Right) != 1 {
		t.Error("expected one intersect operation")
	}
	// The right side should have a difference
	diff := intersect2.Right[0].Right
	if len(diff.Right) != 1 {
		t.Error("expected difference on right side of intersection")
	}

	// a - b - c should parse as (a - b) - c (left associative)
	expr3 := parser.MustParse("a - b - c")
	diff3 := expr3.Union.Left.Left
	if len(diff3.Right) != 2 {
		t.Errorf("expected two difference operations, got %d", len(diff3.Right))
	}
}

func TestParserFunctionArgs(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	// Test bare identifier argument
	// Note: bare identifiers are parsed as expressions in the AST
	// but the evaluator treats them as identifier arguments
	expr1 := parser.MustParse("func(arg)")
	call := expr1.Union.Left.Left.Left.FuncCall
	if call == nil {
		t.Fatal("expected function call")
	}
	if len(call.Args) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(call.Args))
	}
	// In the AST, it's an expression
	if call.Args[0].Expr == nil {
		t.Fatal("expected expression argument in AST")
	}
	// Verify it's a bare identifier
	if call.Args[0].Expr.Union.Left.Left.Left.Ident == nil ||
		*call.Args[0].Expr.Union.Left.Left.Left.Ident != "arg" {
		t.Error("expected identifier 'arg' in expression")
	}

	// Test string argument
	expr2 := parser.MustParse(`func("string arg")`)
	call2 := expr2.Union.Left.Left.Left.FuncCall
	if call2 == nil {
		t.Fatal("expected function call")
	}
	if len(call2.Args) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(call2.Args))
	}
	if call2.Args[0].StrLit == nil || *call2.Args[0].StrLit != `"string arg"` {
		t.Error("expected string argument")
	}

	// Test expression argument
	expr3 := parser.MustParse("func(a | b)")
	call3 := expr3.Union.Left.Left.Left.FuncCall
	if call3 == nil {
		t.Fatal("expected function call")
	}
	if len(call3.Args) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(call3.Args))
	}
	if call3.Args[0].Expr == nil {
		t.Error("expected expression argument")
	}

	// Test mixed arguments
	expr4 := parser.MustParse(`func(a, "string", b | c)`)
	call4 := expr4.Union.Left.Left.Left.FuncCall
	if call4 == nil {
		t.Fatal("expected function call")
	}
	if len(call4.Args) != 3 {
		t.Fatalf("expected 3 arguments, got %d", len(call4.Args))
	}
	// First arg 'a' is an expression (bare identifier)
	if call4.Args[0].Expr == nil {
		t.Error("expected first argument to be expression")
	}
	// Second arg is a string literal
	if call4.Args[1].StrLit == nil {
		t.Error("expected second argument to be string")
	}
	// Third arg 'b | c' is an expression
	if call4.Args[2].Expr == nil {
		t.Error("expected third argument to be expression")
	}
}

func TestParserWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "spaces around operators",
			input: "a | b & c",
			want:  "a | b & c",
		},
		{
			name:  "no spaces",
			input: "a|b&c",
			want:  "a | b & c",
		},
		{
			name:  "extra spaces",
			input: "  a  |  b  &  c  ",
			want:  "a | b & c",
		},
		{
			name:  "newlines",
			input: "a |\nb &\nc",
			want:  "a | b & c",
		},
		{
			name:  "tabs",
			input: "a\t|\tb\t&\tc",
			want:  "a | b & c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			got := expr.String()
			if got != tt.want {
				t.Errorf("Parse() output = %q, want %q", got, tt.want)
			}
		})
	}
}
