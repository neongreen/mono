package setlang

import (
	"fmt"
	"strings"
)

// Evaluator evaluates set expressions using a provided context.
type Evaluator[T comparable] struct {
	ctx Context[T]
}

// NewEvaluator creates a new evaluator with the given context.
func NewEvaluator[T comparable](ctx Context[T]) *Evaluator[T] {
	return &Evaluator[T]{ctx: ctx}
}

// Eval evaluates an expression and returns the resulting set.
func (e *Evaluator[T]) Eval(expr *Expr) (*Set[T], error) {
	if expr == nil || expr.Union == nil {
		return NewSet[T](), nil
	}
	return e.evalUnion(expr.Union)
}

// evalUnion evaluates a union expression.
func (e *Evaluator[T]) evalUnion(expr *UnionExpr) (*Set[T], error) {
	result, err := e.evalIntersect(expr.Left)
	if err != nil {
		return nil, err
	}

	for _, tail := range expr.Right {
		right, err := e.evalIntersect(tail.Right)
		if err != nil {
			return nil, err
		}
		result = result.Union(right)
	}

	return result, nil
}

// evalIntersect evaluates an intersection expression.
func (e *Evaluator[T]) evalIntersect(expr *IntersectExpr) (*Set[T], error) {
	result, err := e.evalDiff(expr.Left)
	if err != nil {
		return nil, err
	}

	for _, tail := range expr.Right {
		right, err := e.evalDiff(tail.Right)
		if err != nil {
			return nil, err
		}
		result = result.Intersect(right)
	}

	return result, nil
}

// evalDiff evaluates a difference expression.
func (e *Evaluator[T]) evalDiff(expr *DiffExpr) (*Set[T], error) {
	result, err := e.evalPrimary(expr.Left)
	if err != nil {
		return nil, err
	}

	for _, tail := range expr.Right {
		right, err := e.evalPrimary(tail.Right)
		if err != nil {
			return nil, err
		}
		result = result.Diff(right)
	}

	return result, nil
}

// evalPrimary evaluates a primary expression.
func (e *Evaluator[T]) evalPrimary(expr *Primary) (*Set[T], error) {
	if expr.SubExpr != nil {
		return e.Eval(expr.SubExpr)
	}

	if expr.Ident != nil {
		return e.ctx.LookupIdent(*expr.Ident)
	}

	if expr.FuncCall != nil {
		return e.evalFuncCall(expr.FuncCall)
	}

	return nil, fmt.Errorf("invalid primary expression")
}

// evalFuncCall evaluates a function call.
func (e *Evaluator[T]) evalFuncCall(call *FuncCall) (*Set[T], error) {
	// Evaluate arguments
	args := make([]FuncArg[T], len(call.Args))
	for i, arg := range call.Args {
		funcArg, err := e.evalArg(arg)
		if err != nil {
			return nil, fmt.Errorf("error evaluating argument %d of %s: %w", i, call.Name, err)
		}
		args[i] = funcArg
	}

	// Call the function through the context
	result, err := e.ctx.CallFunc(call.Name, args)
	if err != nil {
		return nil, fmt.Errorf("error calling function %s: %w", call.Name, err)
	}

	return result, nil
}

// evalArg evaluates a function argument.
func (e *Evaluator[T]) evalArg(arg *Arg) (FuncArg[T], error) {
	if arg.StrLit != nil {
		// Remove quotes from string literal
		s := *arg.StrLit
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			s = s[1 : len(s)-1]
			// Unescape
			s = strings.ReplaceAll(s, `\"`, `"`)
			s = strings.ReplaceAll(s, `\\`, `\`)
		}
		return FuncArg[T]{StrVal: &s}, nil
	}

	if arg.Ident != nil {
		return FuncArg[T]{Ident: arg.Ident}, nil
	}

	if arg.Expr != nil {
		// Special case: if the expression is just a bare identifier, treat it as an identifier argument
		// This allows functions to receive identifier names without forcing evaluation
		if isBareIdentifier(arg.Expr) {
			ident := arg.Expr.Union.Left.Left.Left.Ident
			return FuncArg[T]{Ident: ident}, nil
		}

		// Otherwise, evaluate the expression
		set, err := e.Eval(arg.Expr)
		if err != nil {
			return FuncArg[T]{}, err
		}
		return FuncArg[T]{Set: set}, nil
	}

	return FuncArg[T]{}, fmt.Errorf("invalid argument")
}

// isBareIdentifier checks if an expression is just a bare identifier with no operations
func isBareIdentifier(expr *Expr) bool {
	if expr == nil || expr.Union == nil {
		return false
	}
	// Check if there are any union operations
	if len(expr.Union.Right) > 0 {
		return false
	}
	intersect := expr.Union.Left
	if intersect == nil {
		return false
	}
	// Check if there are any intersect operations
	if len(intersect.Right) > 0 {
		return false
	}
	diff := intersect.Left
	if diff == nil {
		return false
	}
	// Check if there are any diff operations
	if len(diff.Right) > 0 {
		return false
	}
	primary := diff.Left
	if primary == nil {
		return false
	}
	// Check if it's just an identifier (not a function call or sub-expression)
	return primary.Ident != nil
}

// Eval is a convenience function that parses and evaluates an expression.
func Eval[T comparable](ctx Context[T], input string) (*Set[T], error) {
	expr, err := Parse(input)
	if err != nil {
		return nil, err
	}

	eval := NewEvaluator(ctx)
	return eval.Eval(expr)
}
