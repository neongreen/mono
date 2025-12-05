// Package tsparser provides pure-Go parsing of TypeScript and TSX source code
// using tree-sitter grammars compiled to WebAssembly and executed via wazero.
//
// The library embeds a pre-built WASM file and requires no CGO, external tools,
// or network access at runtime.
package tsparser

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed internal/wasm/parser.wasm
var parserWASM []byte

// Language represents a parsed language type.
type Language int

const (
	// TypeScript is the TypeScript language.
	TypeScript Language = iota
	// TSX is the TSX language (TypeScript with JSX).
	TSX
)

// tsNodeSize is the size of TSNode struct in bytes.
// TSNode consists of: context[4] (4x4=16 bytes) + id (4 bytes) + tree (4 bytes) = 24 bytes
const tsNodeSize = 24

// Parser manages the WASM runtime and compiled module.
type Parser struct {
	mu      sync.Mutex
	ctx     context.Context
	runtime wazero.Runtime
	module  wazero.CompiledModule
}

// Tree represents a parsed syntax tree.
type Tree struct {
	parser   *Parser
	module   api.Module
	treePtr  uint64
	src      []byte
	language Language
}

// Node represents a node in the syntax tree.
type Node struct {
	tree    *Tree
	nodePtr uint64
}

// NewParser creates a new parser with compiled WASM modules.
// The parser should be closed with Close() when no longer needed.
func NewParser(ctx context.Context) (*Parser, error) {
	r := wazero.NewRuntime(ctx)

	// Instantiate WASI
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	// Compile the combined parser module (contains both TypeScript and TSX grammars)
	module, err := r.CompileModule(ctx, parserWASM)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to compile parser WASM: %w", err)
	}

	return &Parser{
		ctx:     ctx,
		runtime: r,
		module:  module,
	}, nil
}

// ParseTS parses TypeScript source code and returns a syntax tree.
func (p *Parser) ParseTS(src []byte) (*Tree, error) {
	return p.parse(src, TypeScript)
}

// ParseTSX parses TSX source code and returns a syntax tree.
func (p *Parser) ParseTSX(src []byte) (*Tree, error) {
	return p.parse(src, TSX)
}

func (p *Parser) parse(src []byte, lang Language) (*Tree, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var langFuncName string

	switch lang {
	case TypeScript:
		langFuncName = "tree_sitter_typescript"
	case TSX:
		langFuncName = "tree_sitter_tsx"
	default:
		return nil, fmt.Errorf("unknown language: %d", lang)
	}

	// Instantiate the module
	config := wazero.NewModuleConfig().WithName("")
	module, err := p.runtime.InstantiateModule(p.ctx, p.module, config)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}

	// Get required functions
	malloc := module.ExportedFunction("malloc")
	parserNew := module.ExportedFunction("ts_parser_new")
	parserSetLanguage := module.ExportedFunction("ts_parser_set_language")
	parserParseString := module.ExportedFunction("ts_parser_parse_string")
	parserDelete := module.ExportedFunction("ts_parser_delete")
	languageFunc := module.ExportedFunction(langFuncName)

	if malloc == nil || parserNew == nil || parserSetLanguage == nil ||
		parserParseString == nil || parserDelete == nil || languageFunc == nil {
		module.Close(p.ctx)
		return nil, fmt.Errorf("missing required WASM exports")
	}

	// Get language pointer
	langResult, err := languageFunc.Call(p.ctx)
	if err != nil {
		module.Close(p.ctx)
		return nil, fmt.Errorf("failed to get language: %w", err)
	}
	langPtr := langResult[0]

	// Create parser
	parserResult, err := parserNew.Call(p.ctx)
	if err != nil {
		module.Close(p.ctx)
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}
	parserPtr := parserResult[0]

	// Set language
	setLangResult, err := parserSetLanguage.Call(p.ctx, parserPtr, langPtr)
	if err != nil {
		parserDelete.Call(p.ctx, parserPtr)
		module.Close(p.ctx)
		return nil, fmt.Errorf("failed to set language: %w", err)
	}
	if setLangResult[0] == 0 {
		parserDelete.Call(p.ctx, parserPtr)
		module.Close(p.ctx)
		return nil, fmt.Errorf("incompatible language version")
	}

	// Allocate and copy source string to WASM memory
	srcSize := uint64(len(src))
	mallocResult, err := malloc.Call(p.ctx, srcSize+1) // +1 for null terminator
	if err != nil {
		parserDelete.Call(p.ctx, parserPtr)
		module.Close(p.ctx)
		return nil, fmt.Errorf("failed to allocate memory: %w", err)
	}
	srcPtr := mallocResult[0]

	// Write source to memory
	if !module.Memory().Write(uint32(srcPtr), src) {
		parserDelete.Call(p.ctx, parserPtr)
		module.Close(p.ctx)
		return nil, fmt.Errorf("failed to write source to memory")
	}
	// Write null terminator
	if !module.Memory().WriteByte(uint32(srcPtr+srcSize), 0) {
		parserDelete.Call(p.ctx, parserPtr)
		module.Close(p.ctx)
		return nil, fmt.Errorf("failed to write null terminator")
	}

	// Parse the source
	// ts_parser_parse_string(parser, old_tree, string, length) -> tree
	parseResult, err := parserParseString.Call(p.ctx, parserPtr, 0, srcPtr, srcSize)

	// Free the source memory now that parsing is complete
	freeFunc := module.ExportedFunction("free")
	if freeFunc != nil {
		freeFunc.Call(p.ctx, srcPtr)
	}

	if err != nil {
		parserDelete.Call(p.ctx, parserPtr)
		module.Close(p.ctx)
		return nil, fmt.Errorf("failed to parse: %w", err)
	}
	treePtr := parseResult[0]

	if treePtr == 0 {
		parserDelete.Call(p.ctx, parserPtr)
		module.Close(p.ctx)
		return nil, fmt.Errorf("parsing failed: null tree returned")
	}

	// Delete parser (tree is still valid)
	parserDelete.Call(p.ctx, parserPtr)

	return &Tree{
		parser:   p,
		module:   module,
		treePtr:  treePtr,
		src:      src,
		language: lang,
	}, nil
}

// Close releases all resources used by the parser.
func (p *Parser) Close(ctx context.Context) error {
	return p.runtime.Close(ctx)
}

// RootNode returns the root node of the syntax tree.
func (t *Tree) RootNode() (*Node, error) {
	// TSNode is a 24-byte struct: context[4] (4x4=16 bytes) + id (4 bytes) + tree (4 bytes)
	// We need to allocate space for the result
	malloc := t.module.ExportedFunction("malloc")
	treeRootNode := t.module.ExportedFunction("ts_tree_root_node")

	if malloc == nil || treeRootNode == nil {
		return nil, fmt.Errorf("missing required WASM exports")
	}

	// Allocate space for TSNode (24 bytes)
	nodeResult, err := malloc.Call(t.parser.ctx, tsNodeSize)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate node: %w", err)
	}
	nodePtr := nodeResult[0]

	// Call ts_tree_root_node(nodePtr, treePtr) - returns node via pointer
	_, err = treeRootNode.Call(t.parser.ctx, nodePtr, t.treePtr)
	if err != nil {
		return nil, fmt.Errorf("failed to get root node: %w", err)
	}

	return &Node{
		tree:    t,
		nodePtr: nodePtr,
	}, nil
}

// Close releases resources used by the tree.
func (t *Tree) Close() error {
	treeDelete := t.module.ExportedFunction("ts_tree_delete")
	if treeDelete != nil {
		treeDelete.Call(t.parser.ctx, t.treePtr)
	}
	return t.module.Close(t.parser.ctx)
}

// Source returns the original source code.
func (t *Tree) Source() []byte {
	return t.src
}

// TypeName returns the type name of this node.
func (n *Node) TypeName() (string, error) {
	nodeType := n.tree.module.ExportedFunction("ts_node_type")
	if nodeType == nil {
		return "", fmt.Errorf("missing ts_node_type export")
	}

	result, err := nodeType.Call(n.tree.parser.ctx, n.nodePtr)
	if err != nil {
		return "", fmt.Errorf("failed to get node type: %w", err)
	}

	return n.readString(result[0])
}

// StartByte returns the byte offset where this node starts.
func (n *Node) StartByte() (uint32, error) {
	nodeStartByte := n.tree.module.ExportedFunction("ts_node_start_byte")
	if nodeStartByte == nil {
		return 0, fmt.Errorf("missing ts_node_start_byte export")
	}

	result, err := nodeStartByte.Call(n.tree.parser.ctx, n.nodePtr)
	if err != nil {
		return 0, fmt.Errorf("failed to get start byte: %w", err)
	}

	return uint32(result[0]), nil
}

// EndByte returns the byte offset where this node ends.
func (n *Node) EndByte() (uint32, error) {
	nodeEndByte := n.tree.module.ExportedFunction("ts_node_end_byte")
	if nodeEndByte == nil {
		return 0, fmt.Errorf("missing ts_node_end_byte export")
	}

	result, err := nodeEndByte.Call(n.tree.parser.ctx, n.nodePtr)
	if err != nil {
		return 0, fmt.Errorf("failed to get end byte: %w", err)
	}

	return uint32(result[0]), nil
}

// ChildCount returns the number of children of this node.
func (n *Node) ChildCount() (uint32, error) {
	nodeChildCount := n.tree.module.ExportedFunction("ts_node_child_count")
	if nodeChildCount == nil {
		return 0, fmt.Errorf("missing ts_node_child_count export")
	}

	result, err := nodeChildCount.Call(n.tree.parser.ctx, n.nodePtr)
	if err != nil {
		return 0, fmt.Errorf("failed to get child count: %w", err)
	}

	return uint32(result[0]), nil
}

// NamedChildCount returns the number of named children of this node.
func (n *Node) NamedChildCount() (uint32, error) {
	nodeNamedChildCount := n.tree.module.ExportedFunction("ts_node_named_child_count")
	if nodeNamedChildCount == nil {
		return 0, fmt.Errorf("missing ts_node_named_child_count export")
	}

	result, err := nodeNamedChildCount.Call(n.tree.parser.ctx, n.nodePtr)
	if err != nil {
		return 0, fmt.Errorf("failed to get named child count: %w", err)
	}

	return uint32(result[0]), nil
}

// Child returns the child at the given index.
func (n *Node) Child(index uint32) (*Node, error) {
	malloc := n.tree.module.ExportedFunction("malloc")
	nodeChild := n.tree.module.ExportedFunction("ts_node_child")

	if malloc == nil || nodeChild == nil {
		return nil, fmt.Errorf("missing required WASM exports")
	}

	// Allocate space for TSNode (24 bytes)
	mallocResult, err := malloc.Call(n.tree.parser.ctx, tsNodeSize)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate node: %w", err)
	}
	childPtr := mallocResult[0]

	// ts_node_child(childPtr, nodePtr, index)
	_, err = nodeChild.Call(n.tree.parser.ctx, childPtr, n.nodePtr, uint64(index))
	if err != nil {
		return nil, fmt.Errorf("failed to get child: %w", err)
	}

	return &Node{
		tree:    n.tree,
		nodePtr: childPtr,
	}, nil
}

// NamedChild returns the named child at the given index.
func (n *Node) NamedChild(index uint32) (*Node, error) {
	malloc := n.tree.module.ExportedFunction("malloc")
	nodeNamedChild := n.tree.module.ExportedFunction("ts_node_named_child")

	if malloc == nil || nodeNamedChild == nil {
		return nil, fmt.Errorf("missing required WASM exports")
	}

	// Allocate space for TSNode (24 bytes)
	mallocResult, err := malloc.Call(n.tree.parser.ctx, tsNodeSize)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate node: %w", err)
	}
	childPtr := mallocResult[0]

	// ts_node_named_child(childPtr, nodePtr, index)
	_, err = nodeNamedChild.Call(n.tree.parser.ctx, childPtr, n.nodePtr, uint64(index))
	if err != nil {
		return nil, fmt.Errorf("failed to get named child: %w", err)
	}

	return &Node{
		tree:    n.tree,
		nodePtr: childPtr,
	}, nil
}

// IsNull returns true if this is a null node.
func (n *Node) IsNull() (bool, error) {
	nodeIsNull := n.tree.module.ExportedFunction("ts_node_is_null")
	if nodeIsNull == nil {
		return false, fmt.Errorf("missing ts_node_is_null export")
	}

	result, err := nodeIsNull.Call(n.tree.parser.ctx, n.nodePtr)
	if err != nil {
		return false, fmt.Errorf("failed to check if null: %w", err)
	}

	return result[0] != 0, nil
}

// IsError returns true if this node represents a syntax error.
func (n *Node) IsError() (bool, error) {
	nodeIsError := n.tree.module.ExportedFunction("ts_node_is_error")
	if nodeIsError == nil {
		return false, fmt.Errorf("missing ts_node_is_error export")
	}

	result, err := nodeIsError.Call(n.tree.parser.ctx, n.nodePtr)
	if err != nil {
		return false, fmt.Errorf("failed to check if error: %w", err)
	}

	return result[0] != 0, nil
}

// Parent returns the parent of this node.
func (n *Node) Parent() (*Node, error) {
	malloc := n.tree.module.ExportedFunction("malloc")
	nodeParent := n.tree.module.ExportedFunction("ts_node_parent")

	if malloc == nil || nodeParent == nil {
		return nil, fmt.Errorf("missing required WASM exports")
	}

	// Allocate space for TSNode (24 bytes)
	mallocResult, err := malloc.Call(n.tree.parser.ctx, tsNodeSize)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate node: %w", err)
	}
	parentPtr := mallocResult[0]

	// ts_node_parent(parentPtr, nodePtr)
	_, err = nodeParent.Call(n.tree.parser.ctx, parentPtr, n.nodePtr)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent: %w", err)
	}

	return &Node{
		tree:    n.tree,
		nodePtr: parentPtr,
	}, nil
}

// NextSibling returns the next sibling of this node.
func (n *Node) NextSibling() (*Node, error) {
	malloc := n.tree.module.ExportedFunction("malloc")
	nodeNextSibling := n.tree.module.ExportedFunction("ts_node_next_sibling")

	if malloc == nil || nodeNextSibling == nil {
		return nil, fmt.Errorf("missing required WASM exports")
	}

	// Allocate space for TSNode (24 bytes)
	mallocResult, err := malloc.Call(n.tree.parser.ctx, tsNodeSize)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate node: %w", err)
	}
	siblingPtr := mallocResult[0]

	// ts_node_next_sibling(siblingPtr, nodePtr)
	_, err = nodeNextSibling.Call(n.tree.parser.ctx, siblingPtr, n.nodePtr)
	if err != nil {
		return nil, fmt.Errorf("failed to get next sibling: %w", err)
	}

	return &Node{
		tree:    n.tree,
		nodePtr: siblingPtr,
	}, nil
}

// PrevSibling returns the previous sibling of this node.
func (n *Node) PrevSibling() (*Node, error) {
	malloc := n.tree.module.ExportedFunction("malloc")
	nodePrevSibling := n.tree.module.ExportedFunction("ts_node_prev_sibling")

	if malloc == nil || nodePrevSibling == nil {
		return nil, fmt.Errorf("missing required WASM exports")
	}

	// Allocate space for TSNode (24 bytes)
	mallocResult, err := malloc.Call(n.tree.parser.ctx, tsNodeSize)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate node: %w", err)
	}
	siblingPtr := mallocResult[0]

	// ts_node_prev_sibling(siblingPtr, nodePtr)
	_, err = nodePrevSibling.Call(n.tree.parser.ctx, siblingPtr, n.nodePtr)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous sibling: %w", err)
	}

	return &Node{
		tree:    n.tree,
		nodePtr: siblingPtr,
	}, nil
}

// NextNamedSibling returns the next named sibling of this node.
func (n *Node) NextNamedSibling() (*Node, error) {
	malloc := n.tree.module.ExportedFunction("malloc")
	nodeNextNamedSibling := n.tree.module.ExportedFunction("ts_node_next_named_sibling")

	if malloc == nil || nodeNextNamedSibling == nil {
		return nil, fmt.Errorf("missing required WASM exports")
	}

	// Allocate space for TSNode (24 bytes)
	mallocResult, err := malloc.Call(n.tree.parser.ctx, tsNodeSize)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate node: %w", err)
	}
	siblingPtr := mallocResult[0]

	// ts_node_next_named_sibling(siblingPtr, nodePtr)
	_, err = nodeNextNamedSibling.Call(n.tree.parser.ctx, siblingPtr, n.nodePtr)
	if err != nil {
		return nil, fmt.Errorf("failed to get next named sibling: %w", err)
	}

	return &Node{
		tree:    n.tree,
		nodePtr: siblingPtr,
	}, nil
}

// PrevNamedSibling returns the previous named sibling of this node.
func (n *Node) PrevNamedSibling() (*Node, error) {
	malloc := n.tree.module.ExportedFunction("malloc")
	nodePrevNamedSibling := n.tree.module.ExportedFunction("ts_node_prev_named_sibling")

	if malloc == nil || nodePrevNamedSibling == nil {
		return nil, fmt.Errorf("missing required WASM exports")
	}

	// Allocate space for TSNode (24 bytes)
	mallocResult, err := malloc.Call(n.tree.parser.ctx, tsNodeSize)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate node: %w", err)
	}
	siblingPtr := mallocResult[0]

	// ts_node_prev_named_sibling(siblingPtr, nodePtr)
	_, err = nodePrevNamedSibling.Call(n.tree.parser.ctx, siblingPtr, n.nodePtr)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous named sibling: %w", err)
	}

	return &Node{
		tree:    n.tree,
		nodePtr: siblingPtr,
	}, nil
}

// Text returns the source text corresponding to this node.
func (n *Node) Text() (string, error) {
	start, err := n.StartByte()
	if err != nil {
		return "", err
	}
	end, err := n.EndByte()
	if err != nil {
		return "", err
	}

	src := n.tree.Source()
	if int(end) > len(src) {
		end = uint32(len(src))
	}
	if start > end {
		return "", nil
	}

	return string(src[start:end]), nil
}

// readString reads a null-terminated string from WASM memory.
func (n *Node) readString(ptr uint64) (string, error) {
	strlen := n.tree.module.ExportedFunction("strlen")
	if strlen == nil {
		return "", fmt.Errorf("missing strlen export")
	}

	lenResult, err := strlen.Call(n.tree.parser.ctx, ptr)
	if err != nil {
		return "", fmt.Errorf("failed to get string length: %w", err)
	}

	strLen := uint32(lenResult[0])
	strBytes, ok := n.tree.module.Memory().Read(uint32(ptr), strLen)
	if !ok {
		return "", fmt.Errorf("failed to read string from memory")
	}

	return string(strBytes), nil
}
