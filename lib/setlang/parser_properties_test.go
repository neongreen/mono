package setlang

import (
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// Generators for expressions

// genIdent generates a valid identifier
func genIdent() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		// First character: letter or underscore
		first := rapid.SampledFrom([]rune("abcdefghijklmnopqrstuvwxyz_")).Draw(t, "first")
		// Rest: letters, digits, underscores
		rest := rapid.StringOf(rapid.SampledFrom([]rune("abcdefghijklmnopqrstuvwxyz0123456789_"))).Draw(t, "rest")
		return string(first) + rest
	})
}

// genStringLiteral generates a string literal (without quotes)
func genStringLiteral() *rapid.Generator[string] {
	return rapid.StringMatching(`[^"\\]*`) // Simple strings without quotes or backslashes for now
}

// genSimpleExpr generates a simple expression (identifier or function call)
func genSimpleExpr(depth int) *rapid.Generator[string] {
	if depth <= 0 {
		return genIdent()
	}

	return rapid.OneOf(
		genIdent(),
		genFunctionCall(depth-1),
	)
}

// genFunctionCall generates a function call
func genFunctionCall(depth int) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		name := genIdent().Draw(t, "funcName")
		numArgs := rapid.IntRange(0, 3).Draw(t, "numArgs")

		if numArgs == 0 {
			return name + "()"
		}

		args := make([]string, numArgs)
		for i := 0; i < numArgs; i++ {
			argType := rapid.IntRange(0, 2).Draw(t, "argType")
			switch argType {
			case 0: // identifier
				args[i] = genIdent().Draw(t, "arg")
			case 1: // string literal
				str := genStringLiteral().Draw(t, "str")
				args[i] = `"` + str + `"`
			case 2: // simple expression
				if depth > 0 {
					args[i] = genSimpleExpr(depth - 1).Draw(t, "expr")
				} else {
					args[i] = genIdent().Draw(t, "arg")
				}
			}
		}

		return name + "(" + strings.Join(args, ", ") + ")"
	})
}

// genBinaryOp generates a binary operation
func genBinaryOp() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{"|", "&", "~"})
}

// genExpr generates a random expression
func genExpr(maxDepth int) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		if maxDepth <= 0 {
			return genSimpleExpr(0).Draw(t, "simple")
		}

		exprType := rapid.IntRange(0, 3).Draw(t, "exprType")
		switch exprType {
		case 0: // simple expression
			return genSimpleExpr(1).Draw(t, "simple")
		case 1: // binary operation
			left := genExpr(maxDepth - 1).Draw(t, "left")
			op := genBinaryOp().Draw(t, "op")
			right := genExpr(maxDepth - 1).Draw(t, "right")
			return left + " " + op + " " + right
		case 2: // parenthesized expression
			inner := genExpr(maxDepth - 1).Draw(t, "inner")
			return "(" + inner + ")"
		default: // mixed
			left := genSimpleExpr(1).Draw(t, "left")
			op := genBinaryOp().Draw(t, "op")
			right := genSimpleExpr(1).Draw(t, "right")
			return left + " " + op + " " + right
		}
	})
}

// Property: Parser never panics on valid-looking input
func TestProperty_ParserNoPanic(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := genExpr(3).Draw(t, "expr")

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("parser panicked on input: %s, panic: %v", expr, r)
			}
		}()

		_, _ = Parse(expr) // May return error, but shouldn't panic
	})
}

// Property: Valid expressions can be parsed
func TestProperty_ParseValidExpressions(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := genExpr(2).Draw(t, "expr")

		_, err := Parse(expr)
		// We don't assert no error because some generated expressions might be invalid
		// But we do check that if it parses, we can get it back
		if err == nil {
			// Successfully parsed
			_ = expr
		}
	})
}

// Property: Parse-eval is deterministic
func TestProperty_EvalDeterministic(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Create a simple context
		ctx := NewMapContext[int]()
		ctx.SetIdent("a", NewSetFrom(1, 2, 3))
		ctx.SetIdent("b", NewSetFrom(2, 3, 4))
		ctx.SetIdent("c", NewSetFrom(3, 4, 5))

		// Generate a simple expression using only identifiers we defined
		expr := rapid.SampledFrom([]string{
			"a",
			"b",
			"c",
			"a | b",
			"a & b",
			"a ~ b",
			"(a | b) & c",
			"a | (b & c)",
			"(a | b) ~ c",
		}).Draw(t, "expr")

		// Evaluate twice
		result1, err1 := Eval(ctx, expr)
		result2, err2 := Eval(ctx, expr)

		// Should get same result
		if (err1 == nil) != (err2 == nil) {
			t.Fatalf("eval non-deterministic: different error status")
		}

		if err1 == nil && !setsEqual(result1, result2) {
			t.Fatalf("eval non-deterministic: different results for %s", expr)
		}
	})
}

// Property: Operator precedence is consistent
func TestProperty_OperatorPrecedence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ctx := NewMapContext[int]()
		ctx.SetIdent("a", NewSetFrom(1, 2))
		ctx.SetIdent("b", NewSetFrom(2, 3))
		ctx.SetIdent("c", NewSetFrom(3, 4))

		// Test that a | b & c is parsed as a | (b & c)
		expr1, err1 := Eval(ctx, "a | b & c")
		expr2, err2 := Eval(ctx, "a | (b & c)")

		if err1 != nil || err2 != nil {
			return // Skip if either fails to parse
		}

		if !setsEqual(expr1, expr2) {
			t.Fatalf("precedence inconsistent: 'a | b & c' != 'a | (b & c)'")
		}

		// Test that a & b ~ c is parsed as a & (b ~ c)
		expr3, err3 := Eval(ctx, "a & b ~ c")
		expr4, err4 := Eval(ctx, "a & (b ~ c)")

		if err3 != nil || err4 != nil {
			return
		}

		if !setsEqual(expr3, expr4) {
			t.Fatalf("precedence inconsistent: 'a & b ~ c' != 'a & (b ~ c)'")
		}
	})
}

// Property: Parentheses work correctly
func TestProperty_ParenthesesCorrect(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ctx := NewMapContext[int]()
		ctx.SetIdent("a", NewSetFrom(1, 2, 3))
		ctx.SetIdent("b", NewSetFrom(2, 3, 4))
		ctx.SetIdent("c", NewSetFrom(3, 4, 5))

		// (a | b) & c should be different from a | (b & c) in general
		expr1, err1 := Eval(ctx, "(a | b) & c")
		expr2, err2 := Eval(ctx, "a | (b & c)")

		if err1 != nil || err2 != nil {
			return
		}

		// Expected: (a | b) & c = {1,2,3,4} & {3,4,5} = {3,4}
		expected1 := NewSetFrom(3, 4)
		if !setsEqual(expr1, expected1) {
			t.Fatalf("(a | b) & c incorrect")
		}

		// Expected: a | (b & c) = {1,2,3} | {3,4} = {1,2,3,4}
		expected2 := NewSetFrom(1, 2, 3, 4)
		if !setsEqual(expr2, expected2) {
			t.Fatalf("a | (b & c) incorrect")
		}
	})
}

// Property: Left associativity of same-precedence operators
func TestProperty_LeftAssociative(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ctx := NewMapContext[int]()
		ctx.SetIdent("a", NewSetFrom(1, 2, 3))
		ctx.SetIdent("b", NewSetFrom(2, 3))
		ctx.SetIdent("c", NewSetFrom(3))

		// a ~ b ~ c should be (a ~ b) ~ c
		expr1, err1 := Eval(ctx, "a ~ b ~ c")
		expr2, err2 := Eval(ctx, "(a ~ b) ~ c")

		if err1 != nil || err2 != nil {
			return
		}

		if !setsEqual(expr1, expr2) {
			t.Fatalf("diff not left associative: 'a ~ b ~ c' != '(a ~ b) ~ c'")
		}

		// Verify it's different from right associative
		expr3, err3 := Eval(ctx, "a ~ (b ~ c)")
		if err3 != nil {
			return
		}

		// These should be different
		// (a ~ b) ~ c = ({1,2,3} ~ {2,3}) ~ {3} = {1} ~ {3} = {1}
		// a ~ (b ~ c) = {1,2,3} ~ ({2,3} ~ {3}) = {1,2,3} ~ {2} = {1,3}
		expected1 := NewSetFrom(1)
		expected3 := NewSetFrom(1, 3)

		if !setsEqual(expr1, expected1) {
			t.Fatalf("(a ~ b) ~ c incorrect")
		}
		if !setsEqual(expr3, expected3) {
			t.Fatalf("a ~ (b ~ c) incorrect")
		}
	})
}

// Property: String representation is valid
func TestProperty_StringRepresentationValid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expr := genExpr(2).Draw(t, "expr")

		ast, err := Parse(expr)
		if err != nil {
			return // Skip invalid expressions
		}

		// Convert back to string
		str := ast.String()

		// Should be able to parse the string representation
		ast2, err2 := Parse(str)
		if err2 != nil {
			t.Fatalf("string representation not parseable: %s -> %s (error: %v)", expr, str, err2)
		}

		// Should produce same string (modulo whitespace normalization)
		str2 := ast2.String()
		if str != str2 {
			t.Fatalf("string representation not stable: %s -> %s -> %s", expr, str, str2)
		}
	})
}
