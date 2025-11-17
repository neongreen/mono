# pathlang

A domain-agnostic path query language for expressing hierarchical resource locators with predicates, inspired by XPath but designed for simplicity and ease of use.

## Overview

Pathlang provides a simple, expressive syntax for selecting resources in hierarchical structures. It's designed to be:

- **Domain agnostic**: Works with any hierarchical data through a resolver interface
- **Easy to type**: Shell-friendly syntax without excessive quoting
- **Simple but extensible**: Small core syntax with room for future growth
- **Separate from actions**: Pure selection language (actions like `@close` live outside this library)

## Syntax

### Basic Examples

```
/                                                # Root
/projects                                        # All projects
/projects[name=foo]                             # Projects where name equals "foo"
/projects[name="foo bar"]                       # Quoted values for spaces
/projects[name=foo]/tasks[status=open]          # Nested path
/tasks[status=open,assignee=me]                 # Multiple predicates (AND)
/tasks[description~=bug]                        # Pattern matching
```

### Grammar

Paths are always absolute (start with `/`) and consist of segments separated by `/`:

```
Path      ::= "/" Segments?
Segments  ::= Segment ( "/" Segment )*
Segment   ::= Name Predicates?
Predicates ::= "[" PredicateList "]"
PredicateList ::= Predicate ( "," Predicate )*
Predicate ::= FieldName Op Value
```

### Identifiers

Identifiers (segment names and field names) can contain:
- Letters (a-z, A-Z)
- Digits (0-9, but not at the start)
- Underscores (`_`)
- Hyphens (`-`)

Examples: `projects`, `task_items`, `my-field-name`

### Operators

| Operator | Meaning | Example |
|----------|---------|---------|
| `=`      | Equality | `status=open` |
| `!=`     | Inequality | `status!=closed` |
| `~=`     | Pattern match | `description~=bug` |

Pattern match semantics (typically substring or glob) are defined by the resolver.

### Values

Values can be **bare** or **quoted**:

**Bare values** can contain any characters except:
- Whitespace
- `/`, `[`, `]`, `,`, `=`, `!`, `~`, `@`, `"`, `'`

**Quoted values** use double quotes with escape sequences:
- `\"` - literal quote
- `\\` - literal backslash
- `\n` - newline
- `\t` - tab

Examples:
```
name=foo              # Bare value
name="foo bar"        # Quoted value with space
path="C:\\Users"      # Escaped backslash
text="line1\nline2"   # Escaped newline
```

### Predicates

All predicates in a segment are implicitly ANDed together:

```
/tasks[status=open,assignee=me]   # status=open AND assignee=me
```

## API

### Parsing

```go
// Parse a path string
path, err := pathlang.Parse("/projects[name=foo]/tasks")
if err != nil {
    // handle error
}

// Convert back to canonical string form
fmt.Println(path.String())  // "/projects[name=foo]/tasks"
```

### Types

```go
type Path struct {
    Segments []Segment
}

type Segment struct {
    Name       string
    Predicates []Predicate
}

type Predicate struct {
    Field string
    Op    Op        // OpEq, OpNotEq, OpMatch
    Value string    // Always a string lexically
}
```

### Evaluation

To evaluate paths against your domain data, implement the `Resolver` interface:

```go
type Resolver interface {
    // Root returns the logical root node
    Root(ctx context.Context) (Node, error)

    // Children resolves a segment under a parent node
    // Should:
    //   - Filter by segment.Name
    //   - Apply all predicates
    //   - Return all matching child nodes
    Children(ctx context.Context, parent Node, seg Segment) ([]Node, error)
}
```

Then evaluate paths:

```go
// Evaluate from root
nodes, err := pathlang.Eval(ctx, resolver, path)

// Or evaluate from a specific node
nodes, err := pathlang.EvalFrom(ctx, resolver, startNode, path)
```

### Implementing a Resolver

Here's a simple example:

```go
type MyResolver struct {
    // your domain data
}

func (r *MyResolver) Root(ctx context.Context) (pathlang.Node, error) {
    return r.rootObject, nil
}

func (r *MyResolver) Children(ctx context.Context, parent pathlang.Node, seg pathlang.Segment) ([]pathlang.Node, error) {
    obj := parent.(*MyObject)

    var matches []pathlang.Node
    for _, child := range obj.Children {
        // Check if child's type matches segment name
        if child.Type() != seg.Name {
            continue
        }

        // Apply predicates
        if !r.matchesPredicates(child, seg.Predicates) {
            continue
        }

        matches = append(matches, child)
    }

    return matches, nil
}

func (r *MyResolver) matchesPredicates(obj *MyObject, preds []pathlang.Predicate) bool {
    for _, pred := range preds {
        value := obj.GetField(pred.Field)
        switch pred.Op {
        case pathlang.OpEq:
            if value != pred.Value {
                return false
            }
        case pathlang.OpNotEq:
            if value == pred.Value {
                return false
            }
        case pathlang.OpMatch:
            if !strings.Contains(strings.ToLower(value), strings.ToLower(pred.Value)) {
                return false
            }
        }
    }
    return true
}
```

## Semantics

Evaluation is **set-based**:

1. Start with a singleton set containing the root node
2. For each segment:
   - For each node in the current set, get matching children
   - Union all children to form the next set
3. Return the final set

This means paths like `/projects/tasks` will:
1. Start with `{root}`
2. Get all `projects` children → `{project1, project2, ...}`
3. Get all `tasks` children from each project → `{task1, task2, task3, ...}`

## Design Principles

### What's Included (v1)

✅ Absolute paths with hierarchical segments
✅ Simple predicates with AND conjunction
✅ Three operators: `=`, `!=`, `~=`
✅ Quoted and unquoted values
✅ Domain-agnostic resolver interface

### Reserved for Future

These are intentionally not included in v1 but the syntax reserves room for them:

- Boolean operators (`and`, `or`) in predicates
- More comparison operators (`>`, `<`, `>=`, `<=`, `in`)
- Descendant operator (`//tasks[status=open]`)
- Projection/field selection
- Update semantics

The syntax is designed so these extensions won't break existing queries.

## Integration Examples

### With tk (task manager)

```bash
# Pure selection (what pathlang handles)
tk /projects[name=foo]/tasks[status=open]

# With actions (outside pathlang)
tk /tasks[id=43] @close
```

Pathlang only handles the path part; `@close` would be parsed separately by the CLI.

### Custom Application

```go
// Parse user input
path, err := pathlang.Parse(userInput)

// Evaluate against your domain
resolver := &MyDomainResolver{...}
results, err := pathlang.Eval(ctx, resolver, path)

// Process results
for _, node := range results {
    fmt.Printf("Found: %v\n", node)
}
```

## Error Handling

Parse errors include position information:

```go
path, err := pathlang.Parse("/projects[name=")
// Error: parse error at position 15: expected value
//   context: "[name="
//   position:       ^
```

## Testing

Run tests:

```bash
cd lib/pathlang
go test -v
```

The test suite includes:
- Parser tests with valid and invalid inputs
- Round-trip tests (parse → string → parse)
- Evaluation tests with a mock resolver
- Error handling tests

## License

See the LICENSE file in the repository root.
