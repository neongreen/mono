package setlang

import "fmt"

// Context provides the evaluation context for set expressions.
// Consumers of the library implement this interface to provide
// application-specific behavior for identifier lookup and function calls.
type Context[T comparable] interface {
	// LookupIdent resolves an identifier to a set.
	// Returns an error if the identifier is unknown.
	LookupIdent(name string) (*Set[T], error)

	// CallFunc calls a function with the given arguments.
	// Args can be either string literals or evaluated sets.
	// Returns an error if the function is unknown or arguments are invalid.
	CallFunc(name string, args []FuncArg[T]) (*Set[T], error)
}

// FuncArg represents an argument to a function call.
// It can be either a string literal or an evaluated set expression.
type FuncArg[T comparable] struct {
	// StrVal is non-nil if the argument is a string literal
	StrVal *string

	// Ident is non-nil if the argument is an identifier
	Ident *string

	// Set is non-nil if the argument is an evaluated set expression
	Set *Set[T]
}

// IsString returns true if this argument is a string literal.
func (a FuncArg[T]) IsString() bool {
	return a.StrVal != nil
}

// IsIdent returns true if this argument is an identifier.
func (a FuncArg[T]) IsIdent() bool {
	return a.Ident != nil
}

// IsSet returns true if this argument is an evaluated set expression.
func (a FuncArg[T]) IsSet() bool {
	return a.Set != nil
}

// GetString returns the string value, or an error if not a string.
func (a FuncArg[T]) GetString() (string, error) {
	if a.StrVal != nil {
		return *a.StrVal, nil
	}
	return "", fmt.Errorf("argument is not a string")
}

// GetIdent returns the identifier value, or an error if not an identifier.
func (a FuncArg[T]) GetIdent() (string, error) {
	if a.Ident != nil {
		return *a.Ident, nil
	}
	return "", fmt.Errorf("argument is not an identifier")
}

// GetSet returns the set value, or an error if not a set.
func (a FuncArg[T]) GetSet() (*Set[T], error) {
	if a.Set != nil {
		return a.Set, nil
	}
	return nil, fmt.Errorf("argument is not a set expression")
}

// MapContext is a simple implementation of Context that uses maps for lookups.
// This is useful for testing and simple use cases.
type MapContext[T comparable] struct {
	Idents map[string]*Set[T]
	Funcs  map[string]func([]FuncArg[T]) (*Set[T], error)
}

// NewMapContext creates a new MapContext.
func NewMapContext[T comparable]() *MapContext[T] {
	return &MapContext[T]{
		Idents: make(map[string]*Set[T]),
		Funcs:  make(map[string]func([]FuncArg[T]) (*Set[T], error)),
	}
}

// LookupIdent implements Context.LookupIdent.
func (m *MapContext[T]) LookupIdent(name string) (*Set[T], error) {
	set, ok := m.Idents[name]
	if !ok {
		return nil, fmt.Errorf("unknown identifier: %s", name)
	}
	return set, nil
}

// CallFunc implements Context.CallFunc.
func (m *MapContext[T]) CallFunc(name string, args []FuncArg[T]) (*Set[T], error) {
	fn, ok := m.Funcs[name]
	if !ok {
		return nil, fmt.Errorf("unknown function: %s", name)
	}
	return fn(args)
}

// SetIdent registers an identifier with a set.
func (m *MapContext[T]) SetIdent(name string, set *Set[T]) {
	m.Idents[name] = set
}

// SetFunc registers a function.
func (m *MapContext[T]) SetFunc(name string, fn func([]FuncArg[T]) (*Set[T], error)) {
	m.Funcs[name] = fn
}
