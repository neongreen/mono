# Comparison with Jujutsu's Revset/Fileset Languages

This document compares `setlang` with [Jujutsu (jj)](https://github.com/martinvonz/jj)'s revset and fileset languages, analyzing what features we have, what we're missing, and what would be needed to use this library in a JJ clone.

## Overview

Jujutsu uses set-based query languages to select commits (revsets) and files (filesets). Our `setlang` library was inspired by these languages but designed as a generic, domain-agnostic library.

## Current Feature Comparison

### ✅ Features We Have

| Feature | JJ Syntax | setlang Syntax | Notes |
|---------|-----------|----------------|-------|
| Union | `a \| b` | `a \| b` | ✅ Same |
| Intersection | `a & b` | `a & b` | ✅ Same |
| Difference | `a ~ b` | `a - b` | Different operator symbol |
| Parentheses | `(expr)` | `(expr)` | ✅ Same |
| Function calls | `func(arg)` | `func(arg)` | ✅ Same |
| String literals | `"string"` | `"string"` | ✅ Same |
| Identifiers | `name` | `name` | ✅ Same |

### ❌ Features We're Missing

#### 1. **Range Operators** (Critical for JJ)

JJ has range operators that are fundamental to version control:

```jj
# JJ range syntax
x..y          # Ancestors of y that aren't ancestors of x
x..           # Everything after x
..y           # Everything up to y
x:y           # DAG range (x and y and everything between)
```

**Impact**: This is the **most critical missing feature** for a JJ clone. Commit ranges are core to VCS queries.

**Solution**: We would need to add:
- Binary range operators (`..` and `:`)
- Special parsing for these operators
- Context methods to resolve ranges (requires DAG knowledge)

#### 2. **Postfix Operators**

JJ uses postfix operators for ancestry:

```jj
x+            # Descendants of x
x-            # Ancestors of x (conflicts with our binary - operator!)
```

**Impact**: High - ancestry queries are very common.

**Conflict**: Our `-` is binary infix (difference), JJ's `-` is postfix (ancestors). These are incompatible!

**Solution**:
- Add postfix operator support to grammar
- Use different operator names, or
- Disambiguate based on context (complex!)

#### 3. **Unary Negation/Complement**

JJ uses `~` as both unary (complement) and binary (difference):

```jj
~x            # All commits except x (unary)
x ~ y         # Commits in x but not y (binary)
```

**Impact**: Medium - useful but can work around with explicit `all() - x`

**Solution**:
- Add unary operator support to grammar
- Need a concept of "universe" set

#### 4. **String Pattern Prefixes**

JJ supports pattern matching prefixes:

```jj
description(exact:"fix bug")
description(glob:"fix*")
description(regex:"fix.*")
author(substring:"alice")
```

**Impact**: Low - can implement in function handlers

**Solution**: Parse these in the function implementation, not the grammar

### 🤷 Design Differences

#### 1. **Bare Identifiers in Function Arguments**

**setlang**:
```
status(open)     # "open" passed as identifier, not evaluated
```

**JJ**:
```
author(alice)    # Likely evaluated/interpreted by the function
```

**Impact**: Medium - might need different semantics

**Solution**: Functions receive `FuncArg` which includes the identifier name, so functions can decide how to interpret it. Our current design is actually flexible enough.

#### 2. **Built-in vs. Context Functions**

**setlang**: All functions are provided by context (generic)
**JJ**: Many built-in functions (`ancestors()`, `heads()`, etc.)

**Impact**: Low - this is intentional design

**Solution**: N/A - this is a feature. A JJ clone would register these functions in the context.

#### 3. **Operator Precedence**

**setlang**: `-` > `&` > `|`
**JJ**: Different precedence with ranges, ancestry, etc.

**Impact**: Medium - could cause confusion

**Solution**: Would need to adjust if adding JJ's operators

## What Would Break in a JJ Clone?

### Critical Problems

1. **No range operators**: Can't express `main..@` (commits between main and current)
   - **Workaround**: Add as binary operators with custom parsing
   - **Effort**: Medium - grammar changes, precedence rules

2. **Postfix operator conflict**: Our `-` (diff) conflicts with JJ's `-` (ancestors)
   - **Workaround**: Use different operator symbols or rename
   - **Effort**: Low-Medium - grammar change

3. **No ancestry/DAG operators**: Can't traverse commit graphs
   - **Workaround**: Implement as functions: `ancestors(x)`, `descendants(x)`
   - **Effort**: Low - just function implementations
   - **Tradeoff**: More verbose (`ancestors(x)` vs `x-`)

### Medium Problems

1. **No complement operator**: Can't express "everything except X" concisely
   - **Workaround**: Use `all() - x` explicitly
   - **Effort**: Low - just needs `all()` function
   - **Tradeoff**: More verbose

2. **String pattern syntax**: JJ's `exact:`, `glob:`, `regex:` prefixes
   - **Workaround**: Use different function names or parse in function
   - **Effort**: Low - handle in function implementations

### Non-Problems

1. **Domain-specific functions**: JJ's `author()`, `description()`, `branches()`, etc.
   - **Solution**: Implement in context - library is generic by design ✅

2. **Set operations**: Union, intersection, difference
   - **Solution**: Already supported ✅

3. **Function arguments**: Strings, identifiers, expressions
   - **Solution**: Already supported via `FuncArg` ✅

## Recommended Path for JJ Clone

If you were building a JJ clone in Go and wanted to use this library:

### Phase 1: Minimal (Works Today)

Use functions instead of operators:

```go
// Instead of: main..@
// Use:        range(main, @)

// Instead of: @-
// Use:        ancestors(@)

// Instead of: @+
// Use:        descendants(@)

// Instead of: ~x
// Use:        all() - x
```

**Effort**: Low - just implement functions
**Tradeoff**: More verbose, but fully functional

### Phase 2: Add Postfix Operators

Extend grammar to support postfix operators:

```diff
type Primary struct {
    FuncCall *FuncCall `  @@`
    Ident    *string   `| @Ident`
    SubExpr  *Expr     `| "(" @@ ")"`
+   Postfix  *Postfix  `| @@`
}

+type Postfix struct {
+   Base *Primary  `@@`
+   Op   *string   `@("+" | "^")`  // Use ^ instead of - to avoid conflict
+}
```

Map operators:
- `x^` → ancestors (avoiding - conflict)
- `x+` → descendants

**Effort**: Medium
**Benefit**: More JJ-like syntax

### Phase 3: Add Range Operators

Add range operators as binary operations:

```diff
type DiffExpr struct {
    Left  *Primary      `@@`
    Right []*DiffTail   `@@*`
+   Range []*RangeTail  `@@*`
}

+type RangeTail struct {
+   Op    string    `@(".." | ":")`
+   Right *Primary  `@@`
+}
```

**Effort**: Medium-High
**Benefit**: Critical JJ feature

### Phase 4: Add Unary Complement

Add unary negation operator:

```diff
type Primary struct {
+   Not      *Primary  `  "~" @@`
    FuncCall *FuncCall `| @@`
    Ident    *string   `| @Ident`
    SubExpr  *Expr     `| "(" @@ ")"`
}
```

**Effort**: Low-Medium
**Benefit**: More concise "all except X" queries

## Feature Matrix

| Feature | JJ | setlang | Needed for JJ Clone |
|---------|----|---------|--------------------|
| Union (`\|`) | ✅ | ✅ | ✅ Have it |
| Intersection (`&`) | ✅ | ✅ | ✅ Have it |
| Difference | `~` | `-` | ⚠️ Different operator |
| Range (`..`) | ✅ | ❌ | 🔴 Critical - would add |
| DAG range (`:`) | ✅ | ❌ | 🔴 Critical - would add |
| Ancestors postfix | `x-` | ❌ | 🟡 Use `ancestors(x)` |
| Descendants postfix | `x+` | ❌ | 🟡 Use `descendants(x)` |
| Complement | `~x` | ❌ | 🟡 Use `all() - x` |
| Functions | ✅ | ✅ | ✅ Have it |
| String literals | ✅ | ✅ | ✅ Have it |
| Parentheses | ✅ | ✅ | ✅ Have it |
| Pattern prefixes | ✅ | ❌ | 🟢 Handle in functions |

Legend:
- ✅ Have it
- 🟢 Easy workaround (functions)
- 🟡 Workable (more verbose)
- ⚠️ Different but compatible
- 🔴 Would need to add

## Conclusion

### Can You Build a JJ Clone?

**Yes**, with tradeoffs:

1. **Today (Phase 1)**: Fully functional but verbose
   - Use `ancestors(x)` instead of `x-`
   - Use `range(x, y)` instead of `x..y`
   - Use `all() - x` instead of `~x`

2. **With Grammar Extensions (Phase 2-4)**: Near-identical to JJ
   - Add range operators
   - Add postfix operators (with different symbols to avoid conflicts)
   - Add unary complement

### Strengths of Current Design

1. **Generic**: Works for any domain (commits, files, tasks, etc.)
2. **Type-safe**: Go generics provide compile-time safety
3. **Extensible**: Context pattern allows custom functions
4. **Well-tested**: Property tests verify algebraic properties
5. **Clean separation**: Parsing, AST, evaluation are independent

### Weaknesses for JJ Clone

1. **No range operators**: Critical for commit history
2. **Operator conflicts**: `-` means different things
3. **Verbosity**: Must use functions for common operations

### Recommended Approach

For a serious JJ clone, I'd recommend:

1. **Start with current library** for set operations and function evaluation
2. **Fork/extend the grammar** to add:
   - Range operators (`..` and `:`)
   - Postfix operators (use different symbols: `^` for ancestors, `+` for descendants)
   - Unary complement (`~`)
3. **Implement domain functions** in context:
   - `ancestors()`, `descendants()`, `parents()`, `children()`
   - `heads()`, `roots()`, `branches()`, `tags()`
   - `author()`, `description()`, `committer()`, etc.

The library provides a solid foundation, but JJ's VCS-specific operators would need grammar extensions.
