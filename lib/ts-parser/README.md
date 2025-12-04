# ts-parser

A pure-Go library for parsing TypeScript and TSX source code using tree-sitter grammars compiled to WebAssembly and executed via wazero.

## Features

- Pure Go implementation - no CGO required
- Parses TypeScript (`.ts`) and TSX (`.tsx`) files
- Provides syntax tree with byte ranges for each node
- Extracts JSDoc comments and links them to declarations
- Works offline - no network access required
- No external tools needed at runtime

## Installation

```bash
go get github.com/neongreen/mono/lib/ts-parser
```

## Usage

### Parsing TypeScript

```go
package main

import (
    "context"
    "fmt"
    "log"

    tsparser "github.com/neongreen/mono/lib/ts-parser"
)

func main() {
    ctx := context.Background()
    
    // Create a parser
    parser, err := tsparser.NewParser(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer parser.Close(ctx)

    // Parse TypeScript source
    src := []byte(`
function greet(name: string): string {
    return "Hello, " + name;
}
`)

    tree, err := parser.ParseTS(src)
    if err != nil {
        log.Fatal(err)
    }
    defer tree.Close()

    // Get root node
    root, err := tree.RootNode()
    if err != nil {
        log.Fatal(err)
    }

    // Print node information
    typeName, _ := root.TypeName()
    startByte, _ := root.StartByte()
    endByte, _ := root.EndByte()
    
    fmt.Printf("Root: type=%s, bytes=[%d, %d)\n", typeName, startByte, endByte)
    
    // Iterate children
    childCount, _ := root.ChildCount()
    for i := uint32(0); i < childCount; i++ {
        child, _ := root.Child(i)
        childType, _ := child.TypeName()
        text, _ := child.Text()
        fmt.Printf("  Child %d: type=%s, text=%q\n", i, childType, text)
    }
}
```

### Parsing TSX

```go
tree, err := parser.ParseTSX([]byte(`
function Greeting({ name }: { name: string }) {
    return <div>Hello, {name}!</div>;
}
`))
if err != nil {
    log.Fatal(err)
}
defer tree.Close()
```

### Extracting JSDoc Comments

```go
src := []byte(`
/** Greets a person by name */
function greet(name: string): string {
    return "Hello, " + name;
}

class Greeter {
    /** The greeting message */
    message: string;
    
    /** Creates a new Greeter */
    constructor(message: string) {
        this.message = message;
    }
}
`)

tree, err := parser.ParseTS(src)
if err != nil {
    log.Fatal(err)
}
defer tree.Close()

// Extract JSDoc comments
jsdocs, err := tsparser.ExtractJSDoc(src, tree)
if err != nil {
    log.Fatal(err)
}

for _, jsdoc := range jsdocs {
    commentText := string(src[jsdoc.CommentStart:jsdoc.CommentEnd])
    fmt.Printf("JSDoc: %s\n", commentText)
    fmt.Printf("  Attached to: %s (%s)\n", jsdoc.AttachedDecl.Name, jsdoc.AttachedDecl.TypeName)
    fmt.Printf("  Container: %s (%s)\n", jsdoc.ContainerDecl.Name, jsdoc.ContainerDecl.TypeName)
}
```

## API Reference

### Parser

```go
// NewParser creates a new parser with compiled WASM modules.
func NewParser(ctx context.Context) (*Parser, error)

// ParseTS parses TypeScript source code.
func (p *Parser) ParseTS(src []byte) (*Tree, error)

// ParseTSX parses TSX source code.
func (p *Parser) ParseTSX(src []byte) (*Tree, error)

// Close releases all resources.
func (p *Parser) Close(ctx context.Context) error
```

### Tree

```go
// RootNode returns the root node of the syntax tree.
func (t *Tree) RootNode() (*Node, error)

// Close releases resources used by the tree.
func (t *Tree) Close() error

// Source returns the original source code.
func (t *Tree) Source() []byte
```

### Node

```go
// TypeName returns the type name of this node.
func (n *Node) TypeName() (string, error)

// StartByte returns the byte offset where this node starts.
func (n *Node) StartByte() (uint32, error)

// EndByte returns the byte offset where this node ends.
func (n *Node) EndByte() (uint32, error)

// ChildCount returns the number of children.
func (n *Node) ChildCount() (uint32, error)

// Child returns the child at the given index.
func (n *Node) Child(index uint32) (*Node, error)

// NamedChildCount returns the number of named children.
func (n *Node) NamedChildCount() (uint32, error)

// NamedChild returns the named child at the given index.
func (n *Node) NamedChild(index uint32) (*Node, error)

// Parent returns the parent node.
func (n *Node) Parent() (*Node, error)

// NextSibling returns the next sibling.
func (n *Node) NextSibling() (*Node, error)

// PrevSibling returns the previous sibling.
func (n *Node) PrevSibling() (*Node, error)

// Text returns the source text corresponding to this node.
func (n *Node) Text() (string, error)

// IsNull returns true if this is a null node.
func (n *Node) IsNull() (bool, error)

// IsError returns true if this node represents a syntax error.
func (n *Node) IsError() (bool, error)
```

### JSDoc Extraction

```go
// ExtractJSDoc extracts all JSDoc comments from a parsed tree.
func ExtractJSDoc(src []byte, tree *Tree) ([]JSDoc, error)

// JSDoc represents a JSDoc comment with attachment information.
type JSDoc struct {
    CommentStart  int     // Byte offset where comment starts
    CommentEnd    int     // Byte offset where comment ends
    AttachedDecl  NodeRef // Declaration this JSDoc attaches to
    ContainerDecl NodeRef // Nearest enclosing declaration
}

// NodeRef is a stable reference to a node.
type NodeRef struct {
    StartByte int    // Byte offset where node starts
    EndByte   int    // Byte offset where node ends
    Name      string // Display name (may be empty)
    TypeName  string // Grammar type name
}
```

## Building from Source

The WASM files are pre-built and embedded in the library. End users don't need to build them.

For maintainers who need to rebuild the WASM files, see [WASM_BUILD.md](WASM_BUILD.md).

## Constraints

This library is designed to work in restricted environments:

- **CGO_ENABLED=0** - Works without CGO
- **No network access** - Everything is embedded
- **No external tools** - No Node.js, tree-sitter CLI, or compilers needed
- **No C code** - Pure Go with WASM execution via wazero
