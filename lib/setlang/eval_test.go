package setlang

import (
	"fmt"
	"sort"
	"testing"
)

// Helper function to create a test context with some predefined identifiers and functions
func newTestContext() *MapContext[string] {
	ctx := NewMapContext[string]()

	// Add some identifiers
	ctx.SetIdent("a", NewSetFrom("1", "2", "3"))
	ctx.SetIdent("b", NewSetFrom("2", "3", "4"))
	ctx.SetIdent("c", NewSetFrom("3", "4", "5"))
	ctx.SetIdent("empty", NewSet[string]())

	// Add some functions
	ctx.SetFunc("all", func(args []FuncArg[string]) (*Set[string], error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("all() takes no arguments")
		}
		return NewSetFrom("1", "2", "3", "4", "5", "6"), nil
	})

	ctx.SetFunc("single", func(args []FuncArg[string]) (*Set[string], error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("single() takes exactly 1 argument")
		}
		str, err := args[0].GetString()
		if err != nil {
			return nil, err
		}
		return NewSetFrom(str), nil
	})

	ctx.SetFunc("filter", func(args []FuncArg[string]) (*Set[string], error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("filter() takes exactly 2 arguments: a set and a predicate")
		}
		set, err := args[0].GetSet()
		if err != nil {
			return nil, err
		}
		pred, err := args[1].GetString()
		if err != nil {
			return nil, err
		}

		result := NewSet[string]()
		for _, item := range set.Items() {
			if item >= pred { // Simple predicate: keep if >= pred
				result.Add(item)
			}
		}
		return result, nil
	})

	return ctx
}

func TestEvalIdentifier(t *testing.T) {
	ctx := newTestContext()

	result, err := Eval(ctx, "a")
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}

	expected := []string{"1", "2", "3"}
	checkSet(t, result, expected)
}

func TestEvalUnion(t *testing.T) {
	ctx := newTestContext()

	result, err := Eval(ctx, "a | b")
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}

	expected := []string{"1", "2", "3", "4"}
	checkSet(t, result, expected)
}

func TestEvalIntersect(t *testing.T) {
	ctx := newTestContext()

	result, err := Eval(ctx, "a & b")
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}

	expected := []string{"2", "3"}
	checkSet(t, result, expected)
}

func TestEvalDiff(t *testing.T) {
	ctx := newTestContext()

	result, err := Eval(ctx, "a - b")
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}

	expected := []string{"1"}
	checkSet(t, result, expected)
}

func TestEvalPrecedence(t *testing.T) {
	ctx := newTestContext()

	// a | b & c should be a | (b & c)
	// b & c = {3, 4}
	// a | {3, 4} = {1, 2, 3, 4}
	result1, err := Eval(ctx, "a | b & c")
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	expected1 := []string{"1", "2", "3", "4"}
	checkSet(t, result1, expected1)

	// (a | b) & c should be different
	// a | b = {1, 2, 3, 4}
	// {1, 2, 3, 4} & c = {3, 4}
	result2, err := Eval(ctx, "(a | b) & c")
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	expected2 := []string{"3", "4"}
	checkSet(t, result2, expected2)
}

func TestEvalComplexExpression(t *testing.T) {
	ctx := newTestContext()

	// (a | b) & c - b
	// a | b = {1, 2, 3, 4}
	// {1, 2, 3, 4} & c = {3, 4}
	// {3, 4} - b = {}  (since b contains 3 and 4)
	result, err := Eval(ctx, "(a | b) & c - b")
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}

	if !result.IsEmpty() {
		t.Errorf("expected empty set, got %v", result.Items())
	}
}

func TestEvalFunctionCall(t *testing.T) {
	ctx := newTestContext()

	// Test function with no args
	result1, err := Eval(ctx, "all()")
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	expected1 := []string{"1", "2", "3", "4", "5", "6"}
	checkSet(t, result1, expected1)

	// Test function with string arg
	result2, err := Eval(ctx, `single("7")`)
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	expected2 := []string{"7"}
	checkSet(t, result2, expected2)
}

func TestEvalFunctionWithExpressionArg(t *testing.T) {
	ctx := newTestContext()

	// filter(a | b, "2") should keep items >= "2"
	result, err := Eval(ctx, `filter(a | b, "2")`)
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	expected := []string{"2", "3", "4"}
	checkSet(t, result, expected)
}

func TestEvalWithEmpty(t *testing.T) {
	ctx := newTestContext()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "union with empty",
			input:    "a | empty",
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "intersect with empty",
			input:    "a & empty",
			expected: []string{},
		},
		{
			name:     "diff with empty",
			input:    "a - empty",
			expected: []string{"1", "2", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Eval(ctx, tt.input)
			if err != nil {
				t.Fatalf("Eval() error = %v", err)
			}
			checkSet(t, result, tt.expected)
		})
	}
}

func TestEvalErrors(t *testing.T) {
	ctx := newTestContext()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unknown identifier",
			input: "unknown",
		},
		{
			name:  "unknown function",
			input: "unknown_func()",
		},
		{
			name:  "wrong number of arguments",
			input: "single()",
		},
		{
			name:  "wrong argument type",
			input: "single(a | b)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Eval(ctx, tt.input)
			if err == nil {
				t.Error("Eval() expected error, got nil")
			}
		})
	}
}

func TestEvalMultipleOperations(t *testing.T) {
	ctx := newTestContext()

	// Test multiple unions
	result1, err := Eval(ctx, "a | b | c")
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	expected1 := []string{"1", "2", "3", "4", "5"}
	checkSet(t, result1, expected1)

	// Test multiple intersects
	ctx.SetIdent("d", NewSetFrom("2", "3"))
	result2, err := Eval(ctx, "a & b & d")
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	expected2 := []string{"2", "3"}
	checkSet(t, result2, expected2)

	// Test multiple diffs
	result3, err := Eval(ctx, "all() - a - b")
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	// all = {1,2,3,4,5,6}
	// - a = {4,5,6}
	// - b = {5,6}
	expected3 := []string{"5", "6"}
	checkSet(t, result3, expected3)
}

func TestMapContext(t *testing.T) {
	ctx := NewMapContext[int]()

	// Test identifier lookup
	ctx.SetIdent("nums", NewSetFrom(1, 2, 3))
	set, err := ctx.LookupIdent("nums")
	if err != nil {
		t.Fatalf("LookupIdent() error = %v", err)
	}
	if set.Size() != 3 {
		t.Errorf("expected size 3, got %d", set.Size())
	}

	// Test unknown identifier
	_, err = ctx.LookupIdent("unknown")
	if err == nil {
		t.Error("LookupIdent() expected error for unknown identifier")
	}

	// Test function registration and call
	ctx.SetFunc("double", func(args []FuncArg[int]) (*Set[int], error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("double() takes exactly 1 argument")
		}
		set, err := args[0].GetSet()
		if err != nil {
			return nil, err
		}
		result := NewSet[int]()
		for _, item := range set.Items() {
			result.Add(item * 2)
		}
		return result, nil
	})

	// Call through context
	set2 := NewSetFrom(1, 2, 3)
	result, err := ctx.CallFunc("double", []FuncArg[int]{{Set: set2}})
	if err != nil {
		t.Fatalf("CallFunc() error = %v", err)
	}
	expected := []int{2, 4, 6}
	items := result.Items()
	sort.Ints(items)
	for i, v := range expected {
		if items[i] != v {
			t.Errorf("expected %d at index %d, got %d", v, i, items[i])
		}
	}

	// Test unknown function
	_, err = ctx.CallFunc("unknown", nil)
	if err == nil {
		t.Error("CallFunc() expected error for unknown function")
	}
}

func TestFuncArg(t *testing.T) {
	// Test string arg
	str := "test"
	arg1 := FuncArg[int]{StrVal: &str}
	if !arg1.IsString() {
		t.Error("expected IsString() to be true")
	}
	if arg1.IsIdent() || arg1.IsSet() {
		t.Error("expected IsIdent() and IsSet() to be false")
	}
	got1, err := arg1.GetString()
	if err != nil || got1 != "test" {
		t.Errorf("GetString() = %q, %v; want %q, nil", got1, err, "test")
	}

	// Test ident arg
	ident := "myident"
	arg2 := FuncArg[int]{Ident: &ident}
	if !arg2.IsIdent() {
		t.Error("expected IsIdent() to be true")
	}
	if arg2.IsString() || arg2.IsSet() {
		t.Error("expected IsString() and IsSet() to be false")
	}
	got2, err := arg2.GetIdent()
	if err != nil || got2 != "myident" {
		t.Errorf("GetIdent() = %q, %v; want %q, nil", got2, err, "myident")
	}

	// Test set arg
	set := NewSetFrom(1, 2, 3)
	arg3 := FuncArg[int]{Set: set}
	if !arg3.IsSet() {
		t.Error("expected IsSet() to be true")
	}
	if arg3.IsString() || arg3.IsIdent() {
		t.Error("expected IsString() and IsIdent() to be false")
	}
	got3, err := arg3.GetSet()
	if err != nil || got3 != set {
		t.Errorf("GetSet() error or wrong set")
	}

	// Test error cases
	_, err = arg1.GetIdent()
	if err == nil {
		t.Error("expected error when getting ident from string arg")
	}
	_, err = arg2.GetSet()
	if err == nil {
		t.Error("expected error when getting set from ident arg")
	}
	_, err = arg3.GetString()
	if err == nil {
		t.Error("expected error when getting string from set arg")
	}
}

// Helper function to check if a set contains exactly the expected items
func checkSet(t *testing.T, set *Set[string], expected []string) {
	t.Helper()

	if set.Size() != len(expected) {
		t.Errorf("expected size %d, got %d", len(expected), set.Size())
	}

	for _, item := range expected {
		if !set.Has(item) {
			t.Errorf("expected set to contain %q", item)
		}
	}

	items := set.Items()
	if len(items) != len(expected) {
		t.Errorf("Items() returned %d items, expected %d", len(items), len(expected))
	}
}
