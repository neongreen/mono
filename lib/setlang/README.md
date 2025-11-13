# setlang - A Generic Set Expression Language

`setlang` is a standalone Go library for parsing and evaluating set expressions. It provides a flexible, type-safe way to query and manipulate sets using an intuitive expression language inspired by [Jujutsu's](https://github.com/martinvonz/jj) revset and fileset languages.

## Features

- **Generic**: Works with any comparable type (strings, ints, custom types, etc.)
- **Extensible**: Consumer-defined functions for domain-specific operations
- **Type-safe**: Uses Go generics for compile-time type safety
- **Well-tested**: Comprehensive test coverage
- **Fast parsing**: Built on [participle](https://github.com/alecthomas/participle) for efficient parsing
- **Clean API**: Simple Context interface for customization

## Language Syntax

### Operators

- `|` - Union (set A or set B)
- `&` - Intersection (set A and set B)
- `-` - Difference (set A but not set B)

Precedence: `-` (highest) > `&` > `|` (lowest)

### Expressions

```
a | b           # Union of sets a and b
a & b           # Intersection of sets a and b
a - b           # Difference: items in a but not in b
(a | b) & c     # Parentheses for grouping
all()           # Function call with no arguments
status(open)    # Function call with identifier
title("bug")    # Function call with string literal
filter(a | b, "predicate")  # Function call with expression and string
```

### Function Arguments

Function arguments can be:
1. **Bare identifiers**: `status(done)` - passed as identifier name to function
2. **String literals**: `title("hello")` - passed as string value
3. **Set expressions**: `filter(a | b)` - evaluated and passed as a set

Note: Bare identifiers in function arguments are NOT evaluated as expressions. To pass an evaluated set, use a set operation like `filter((a))` or `filter(a | empty())`.

## Installation

```bash
go get github.com/neongreen/mono/lib/setlang
```

## Quick Start

```go
package main

import (
	"fmt"
	"github.com/neongreen/mono/lib/setlang"
)

func main() {
	// Create a context with some predefined sets
	ctx := setlang.NewMapContext[string]()
	ctx.SetIdent("bugs", setlang.NewSetFrom("bug-1", "bug-2", "bug-3"))
	ctx.SetIdent("features", setlang.NewSetFrom("feat-1", "feat-2"))
	ctx.SetIdent("done", setlang.NewSetFrom("bug-1", "feat-1"))

	// Define a function
	ctx.SetFunc("all", func(args []setlang.FuncArg[string]) (*setlang.Set[string], error) {
		return setlang.NewSetFrom("bug-1", "bug-2", "bug-3", "feat-1", "feat-2"), nil
	})

	// Parse and evaluate an expression
	result, err := setlang.Eval(ctx, "(bugs | features) - done")
	if err != nil {
		panic(err)
	}

	// Use the result
	for _, item := range result.Items() {
		fmt.Println(item)
	}
	// Output:
	// bug-2
	// bug-3
	// feat-2
}
```

## Usage

### 1. Define Your Context

Implement the `Context` interface to provide identifier lookup and function implementations:

```go
type Context[T comparable] interface {
	LookupIdent(name string) (*Set[T], error)
	CallFunc(name string, args []FuncArg[T]) (*Set[T], error)
}
```

Or use the provided `MapContext` for simple cases:

```go
ctx := setlang.NewMapContext[MyType]()
ctx.SetIdent("mySet", setlang.NewSetFrom(item1, item2, item3))
ctx.SetFunc("myFunc", func(args []setlang.FuncArg[MyType]) (*setlang.Set[MyType], error) {
	// Your function implementation
	return result, nil
})
```

### 2. Parse Expressions

```go
// Parse an expression into an AST
expr, err := setlang.Parse("a | b & c")
if err != nil {
	// Handle parse error
}
```

### 3. Evaluate Expressions

```go
// Evaluate using a context
eval := setlang.NewEvaluator(ctx)
result, err := eval.Eval(expr)
if err != nil {
	// Handle evaluation error
}

// Or use the convenience function
result, err := setlang.Eval(ctx, "a | b & c")
```

### 4. Work with Sets

```go
// Create sets
set1 := setlang.NewSet[int]()
set2 := setlang.NewSetFrom(1, 2, 3)

// Add/remove items
set1.Add(42)
set1.Remove(42)

// Check membership
if set2.Has(2) {
	// ...
}

// Set operations
union := set1.Union(set2)
intersection := set1.Intersect(set2)
difference := set1.Diff(set2)

// Get items
items := set2.Items() // []int
```

## Implementing Functions

Functions receive arguments as `FuncArg[T]` which can be:
- A string literal (`StrVal`)
- An identifier (`Ident`)
- An evaluated set (`Set`)

Example function that filters by a predicate:

```go
ctx.SetFunc("filter", func(args []setlang.FuncArg[string]) (*setlang.Set[string], error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("filter() takes 2 arguments")
	}

	// First argument: a set expression
	set, err := args[0].GetSet()
	if err != nil {
		return nil, fmt.Errorf("first argument must be a set: %w", err)
	}

	// Second argument: a string predicate
	predicate, err := args[1].GetString()
	if err != nil {
		return nil, fmt.Errorf("second argument must be a string: %w", err)
	}

	// Apply filter
	result := setlang.NewSet[string]()
	for _, item := range set.Items() {
		if matchesPredicate(item, predicate) {
			result.Add(item)
		}
	}
	return result, nil
})
```

Example function that takes an identifier:

```go
ctx.SetFunc("status", func(args []setlang.FuncArg[Task]) (*setlang.Set[Task], error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("status() takes 1 argument")
	}

	// Get the status name (not evaluated as an expression)
	statusName, err := args[0].GetIdent()
	if err != nil {
		// Maybe it's a string literal instead
		statusName, err = args[0].GetString()
		if err != nil {
			return nil, fmt.Errorf("argument must be an identifier or string")
		}
	}

	// Return all tasks with that status
	return getTasksByStatus(statusName), nil
})
```

## Advanced Example

```go
package main

import (
	"fmt"
	"strings"
	"github.com/neongreen/mono/lib/setlang"
)

type Task struct {
	ID     string
	Status string
	Labels []string
}

type TaskContext struct {
	tasks map[string]*Task
}

func (tc *TaskContext) LookupIdent(name string) (*setlang.Set[string], error) {
	// Special identifiers
	switch name {
	case "all":
		result := setlang.NewSet[string]()
		for id := range tc.tasks {
			result.Add(id)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unknown identifier: %s", name)
	}
}

func (tc *TaskContext) CallFunc(name string, args []setlang.FuncArg[string]) (*setlang.Set[string], error) {
	switch name {
	case "status":
		if len(args) != 1 {
			return nil, fmt.Errorf("status() takes 1 argument")
		}
		statusName, _ := args[0].GetIdent()
		if statusName == "" {
			statusName, _ = args[0].GetString()
		}

		result := setlang.NewSet[string]()
		for id, task := range tc.tasks {
			if task.Status == statusName {
				result.Add(id)
			}
		}
		return result, nil

	case "label":
		if len(args) != 1 {
			return nil, fmt.Errorf("label() takes 1 argument")
		}
		labelName, _ := args[0].GetIdent()
		if labelName == "" {
			labelName, _ = args[0].GetString()
		}

		result := setlang.NewSet[string]()
		for id, task := range tc.tasks {
			for _, label := range task.Labels {
				if label == labelName {
					result.Add(id)
					break
				}
			}
		}
		return result, nil

	case "title":
		if len(args) != 1 {
			return nil, fmt.Errorf("title() takes 1 argument")
		}
		pattern, err := args[0].GetString()
		if err != nil {
			return nil, err
		}

		result := setlang.NewSet[string]()
		for id, task := range tc.tasks {
			if strings.Contains(task.ID, pattern) {
				result.Add(id)
			}
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unknown function: %s", name)
	}
}

func main() {
	ctx := &TaskContext{
		tasks: map[string]*Task{
			"task-1": {ID: "task-1", Status: "open", Labels: []string{"bug", "urgent"}},
			"task-2": {ID: "task-2", Status: "open", Labels: []string{"feature"}},
			"task-3": {ID: "task-3", Status: "done", Labels: []string{"bug"}},
		},
	}

	// Find all open bugs
	result, err := setlang.Eval(ctx, "status(open) & label(bug)")
	if err != nil {
		panic(err)
	}

	fmt.Println("Open bugs:", result.Items())
	// Output: Open bugs: [task-1]
}
```

## Testing

The library includes comprehensive tests:

- **Unit tests**: Test individual components and features
- **Property tests**: Verify algebraic properties using [rapid](https://github.com/flyingmutant/rapid)
  - Set operation properties (commutativity, associativity, distributivity)
  - De Morgan's laws
  - Identity and absorption laws
  - Parser/evaluator properties (determinism, precedence, etc.)

Run all tests:

```bash
cd lib/setlang
go test -v
```

Run only property tests:

```bash
go test -v -run TestProperty
```

Property tests run 100 randomized test cases by default, providing high confidence in correctness.

## License

This library is part of the mono repository and follows its license.

## Comparison with Jujutsu

See [JJ_COMPARISON.md](./JJ_COMPARISON.md) for a detailed analysis of how this library compares to Jujutsu's revset/fileset languages, including:
- Feature comparison table
- Missing features and workarounds
- Recommendations for building a JJ clone in Go

## Acknowledgments

Inspired by [Jujutsu's](https://github.com/martinvonz/jj) revset and fileset languages, which provide an elegant way to query version control history.
